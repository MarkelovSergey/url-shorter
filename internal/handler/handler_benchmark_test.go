package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/MarkelovSergey/url-shorter/internal/storage/memorystorage"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "userID"

func setupBenchmarkHandler() *handler {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := urlshorterservice.New(repo, nil, logger)
	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}
	auditPublisher := audit.NewPublisher(logger)

	h := New(cfg, service, nil, logger, auditPublisher)
	return h
}

func BenchmarkCreateHandler(b *testing.B) {
	h := setupBenchmarkHandler()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		body := bytes.NewBufferString("https://example.com/test")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", "text/plain")

		ctx := context.WithValue(req.Context(), userIDKey, "test-user")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateHandler(w, req)
	}
}

func BenchmarkCreateAPIHandler(b *testing.B) {
	h := setupBenchmarkHandler()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reqBody := model.Request{URL: "https://example.com/test"}
		jsonBody, _ := json.Marshal(reqBody)
		body := bytes.NewBuffer(jsonBody)

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		req.Header.Set("Content-Type", "application/json")

		ctx := context.WithValue(req.Context(), userIDKey, "test-user")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateAPIHandler(w, req)
	}
}

func BenchmarkReadHandler(b *testing.B) {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := urlshorterservice.New(repo, nil, logger)

	// Создаём URL напрямую через сервис
	ctx := context.Background()
	shortCode, err := service.Generate(ctx, "https://example.com/test", "test-user")
	if err != nil {
		b.Fatalf("failed to generate URL: %v", err)
	}

	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}
	auditPublisher := audit.NewPublisher(logger)
	h := New(cfg, service, nil, logger, auditPublisher)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/"+shortCode, nil)
		w := httptest.NewRecorder()
		h.ReadHandler(w, req)
	}
}

func BenchmarkCreateBatchHandler(b *testing.B) {
	benchmarks := []struct {
		name      string
		batchSize int
	}{
		{"batch 10", 10},
		{"batch 50", 50},
		{"batch 100", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			h := setupBenchmarkHandler()

			requests := make([]model.BatchRequest, bm.batchSize)
			for i := 0; i < bm.batchSize; i++ {
				requests[i] = model.BatchRequest{
					CorrelationID: "test-id",
					OriginalURL:   "https://example.com/test",
				}
			}
			jsonBody, _ := json.Marshal(requests)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				body := bytes.NewBuffer(jsonBody)
				req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
				req.Header.Set("Content-Type", "application/json")

				ctx := context.WithValue(req.Context(), userIDKey, "test-user")
				req = req.WithContext(ctx)

				w := httptest.NewRecorder()
				h.CreateBatchHandler(w, req)
			}
		})
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	resp := model.Response{Result: "http://localhost:8080/ABC12345"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	jsonData := []byte(`{"url":"https://example.com/test"}`)
	var req model.Request

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(jsonData, &req)
	}
}
