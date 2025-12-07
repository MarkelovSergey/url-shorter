package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type App struct {
	server *http.Server
	dbConn *pgx.Conn
	logger *zap.Logger
}

func New(cfg config.Config) *App {
	var (
		conn       *pgx.Conn
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

		conn, err = pgx.Connect(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Warning: Failed to connect to database: %v", err)
		}

		urlStorage = postgresstorage.New(conn)
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
	healthRepo := healthrepository.New(conn)

	healthService := healthservice.New(healthRepo)
	urlShorterService := urlshorterservice.New(urlShorterRepo, healthRepo)

	handler := handler.New(cfg, urlShorterService, healthService, logger)
	r := chi.NewRouter()
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Gzipping)

	r.Post("/", handler.CreateHandler)
	r.Get("/{id}", handler.ReadHandler)
	r.Post("/api/shorten", handler.CreateAPIHandler)
	r.Get("/ping", handler.PingHandler)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	return &App{
		server: srv,
		dbConn: conn,
		logger: logger,
	}
}

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

	if a.dbConn != nil {
		a.dbConn.Close(context.Background())
	}
	if a.logger != nil {
		a.logger.Sync()
	}

	log.Println("Server exited gracefully")

	return nil
}
