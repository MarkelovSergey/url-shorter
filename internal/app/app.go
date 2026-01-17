package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/handler"
	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/migration"
	"github.com/MarkelovSergey/url-shorter/internal/repository/healthrepository"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
	"github.com/MarkelovSergey/url-shorter/internal/storage/filestorage"
	"github.com/MarkelovSergey/url-shorter/internal/storage/memorystorage"
	"github.com/MarkelovSergey/url-shorter/internal/storage/postgresstorage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// App представляет основное приложение сервиса сокращения URL.
type App struct {
	server         *http.Server
	dbPool         *pgxpool.Pool
	logger         *zap.Logger
	auditPublisher *audit.AuditPublisher
}

// New создает новый экземпляр приложения с заданной конфигурацией.
func New(cfg config.Config) *App {
	var (
		pool       *pgxpool.Pool
		urlStorage storage.Storage
		err        error
	)

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if cfg.DatabaseDSN != "" {
		if err := migration.RunMigrations(cfg.DatabaseDSN); err != nil {
			log.Fatalf("Warning: Failed to run migrations: %v", err)
		}

		pool, err = pgxpool.New(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Warning: Failed to connect to database: %v", err)
		}

		urlStorage = postgresstorage.New(pool)
		log.Println("Using PostgreSQL storage")
	}

	if urlStorage == nil && cfg.FileStoragePath != "" {
		urlStorage = filestorage.New(cfg.FileStoragePath)
		log.Printf("Using file storage: %s", cfg.FileStoragePath)
	}

	if urlStorage == nil {
		urlStorage = memorystorage.New()
		log.Println("Using memory storage")
	}

	urlShorterRepo := urlshorterrepository.New(urlStorage)
	healthRepo := healthrepository.New(pool)

	healthService := healthservice.New(healthRepo)
	urlShorterService := urlshorterservice.New(urlShorterRepo, healthRepo, logger)

	// Инициализация системы аудита
	auditPublisher := audit.NewPublisher(logger)

	if cfg.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(cfg.AuditFile, logger)
		if err != nil {
			log.Printf("Warning: Failed to create file audit observer: %v", err)
		} else {
			auditPublisher.Subscribe(fileObserver)
			log.Printf("Audit file observer enabled: %s", cfg.AuditFile)
		}
	}

	if cfg.AuditURL != "" {
		httpObserver := audit.NewHTTPObserver(cfg.AuditURL, logger)
		auditPublisher.Subscribe(httpObserver)
		log.Printf("Audit HTTP observer enabled: %s", cfg.AuditURL)
	}

	handler := handler.New(cfg, urlShorterService, healthService, logger, auditPublisher)
	r := chi.NewRouter()
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Gzipping)
	r.Use(middleware.Auth)

	r.Post("/", handler.CreateHandler)
	r.Get("/{id}", handler.ReadHandler)
	r.Post("/api/shorten", handler.CreateAPIHandler)
	r.Post("/api/shorten/batch", handler.CreateBatchHandler)
	r.Get("/api/user/urls", handler.GetUserURLsHandler)
	r.Delete("/api/user/urls", handler.DeleteURLsHandler)
	r.Get("/ping", handler.PingHandler)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	return &App{
		server:         srv,
		dbPool:         pool,
		logger:         logger,
		auditPublisher: auditPublisher,
	}
}

// Run запускает HTTP-сервер приложения и ожидает сигнал завершения.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server is starting on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	if a.dbPool != nil {
		a.dbPool.Close()
	}
	if a.auditPublisher != nil {
		a.auditPublisher.Close()
	}
	if a.logger != nil {
		a.logger.Sync()
	}

	log.Println("Server exited gracefully")

	return nil
}
