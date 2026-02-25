package app

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/grpcserver"
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
	pb "github.com/MarkelovSergey/url-shorter/proto"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// App представляет основное приложение сервиса сокращения URL.
// Содержит HTTP-сервер, пул подключений к базе данных,
// логгер и публикатор событий аудита.
type App struct {
	server         *http.Server
	grpcServer     *grpc.Server
	dbPool         *pgxpool.Pool
	logger         *zap.Logger
	auditPublisher *audit.AuditPublisher
	config         config.Config
}

// New создает новый экземпляр приложения с заданной конфигурацией.
// Инициализирует все компоненты: хранилище, сервисы, хендлеры и мидлвары.
// Выполняет миграции базы данных и настраивает систему аудита.
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

	if cfg.Database.DSN != "" {
		if err := migration.RunMigrations(cfg.Database.DSN); err != nil {
			log.Fatalf("Warning: Failed to run migrations: %v", err)
		}

		pool, err = pgxpool.New(context.Background(), cfg.Database.DSN)
		if err != nil {
			log.Fatalf("Warning: Failed to connect to database: %v", err)
		}

		urlStorage = postgresstorage.New(pool)
		log.Println("Using PostgreSQL storage")
	}

	if urlStorage == nil && cfg.Storage.FilePath != "" {
		urlStorage = filestorage.New(cfg.Storage.FilePath)
		log.Printf("Using file storage: %s", cfg.Storage.FilePath)
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

	if cfg.Audit.FilePath != "" {
		fileObserver, err := audit.NewFileObserver(cfg.Audit.FilePath, logger)
		if err != nil {
			log.Printf("Warning: Failed to create file audit observer: %v", err)
		} else {
			auditPublisher.Subscribe(fileObserver)
			log.Printf("Audit file observer enabled: %s", cfg.Audit.FilePath)
		}
	}

	if cfg.Audit.URL != "" {
		httpObserver := audit.NewHTTPObserver(cfg.Audit.URL, logger)
		auditPublisher.Subscribe(httpObserver)
		log.Printf("Audit HTTP observer enabled: %s", cfg.Audit.URL)
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
	r.Get("/api/internal/stats", handler.StatsHandler)

	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: r,
	}

	// Инициализация gRPC-сервера
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.AuthInterceptor()),
	)
	grpcHandler := grpcserver.New(cfg, urlShorterService, logger, auditPublisher)
	pb.RegisterShortenerServiceServer(grpcSrv, grpcHandler)

	return &App{
		server:         srv,
		grpcServer:     grpcSrv,
		dbPool:         pool,
		logger:         logger,
		auditPublisher: auditPublisher,
		config:         cfg,
	}
}

// Run запускает HTTP-сервер приложения и ожидает сигнал завершения.
// Сервер корректно завершается по сигналам SIGINT, SIGTERM или SIGQUIT.
// Закрывает все ресурсы (соединения с БД, логгер, аудит) перед выходом.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	a.server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	go func() {
		if a.config.Server.EnableHTTPS {
			log.Printf("HTTPS server is starting on %s", a.server.Addr)
			if err := a.startTLSServer(); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTPS server failed to start: %v", err)
			}
		} else {
			log.Printf("Server is starting on %s", a.server.Addr)
			if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Server failed to start: %v", err)
			}
		}
	}()

	// Запуск gRPC-сервера
	if a.config.Server.GRPCAddress != "" {
		go func() {
			lis, err := net.Listen("tcp", a.config.Server.GRPCAddress)
			if err != nil {
				log.Printf("Failed to listen for gRPC: %v", err)
				return
			}
			log.Printf("gRPC server is starting on %s", a.config.Server.GRPCAddress)
			if err := a.grpcServer.Serve(lis); err != nil {
				log.Printf("gRPC server failed: %v", err)
			}
		}()
	}

	<-ctx.Done()

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
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

// startTLSServer запускает HTTPS-сервер с самоподписанным сертификатом.
func (a *App) startTLSServer() error {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", a.server.Addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create TLS listener: %w", err)
	}

	return a.server.Serve(listener)
}

// generateSelfSignedCert генерирует самоподписанный TLS сертификат.
func generateSelfSignedCert() (tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"URL Shortener"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	return tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privateKey,
	}, nil
}
