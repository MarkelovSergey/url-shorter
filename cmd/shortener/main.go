package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MarkelovSergey/url-shorter/config"
	"github.com/MarkelovSergey/url-shorter/internal/handler"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"

	"github.com/go-chi/chi/v5"
)

func main() {
	serverAddr := flag.String("a", ":8080", "HTTP server address (e.g. localhost:8888)")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")
	flag.Parse()

	cfg := *config.New(*serverAddr, *baseURL)

	urlShorterRepo := urlshorterrepository.New()
	urlShorterService := urlshorterservice.New(urlShorterRepo)

	handler := handler.New(cfg, urlShorterService)
	r := chi.NewRouter()

	r.Post("/", handler.CreateHandler)
	r.Get("/{id}", handler.ReadHandler)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	go func() {
		log.Printf("Server is starting on %s", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}
