package profiles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/handler"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/MarkelovSergey/url-shorter/internal/storage/memorystorage"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler interface {
	CreateHandler(w http.ResponseWriter, r *http.Request)
	CreateBatchHandler(w http.ResponseWriter, r *http.Request)
	ReadHandler(w http.ResponseWriter, r *http.Request)
}

func setupHandler() Handler {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := urlshorterservice.New(repo, nil, logger)
	cfg := config.Config{
		Server: config.ServerConfig{
			BaseURL: "http://localhost:8080",
		},
	}

	auditPublisher := audit.NewPublisher(logger)

	h := handler.New(cfg, service, nil, logger, auditPublisher)
	return h
}

// runProfiledWorkload выполняет нагрузку для профилирования
func runProfiledWorkload(iterations int) {
	h := setupHandler()
	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	for i := 0; i < iterations; i++ {
		body := bytes.NewBufferString(fmt.Sprintf("https://example.com/test/%d", i))
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", "text/plain")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateHandler(w, req)
	}

	for i := 0; i < iterations/10; i++ {
		requests := make([]model.BatchRequest, 10)
		for j := 0; j < 10; j++ {
			requests[j] = model.BatchRequest{
				CorrelationID: fmt.Sprintf("id-%d-%d", i, j),
				OriginalURL:   fmt.Sprintf("https://batch.example.com/%d/%d", i, j),
			}
		}
		jsonBody, _ := json.Marshal(requests)
		body := bytes.NewBuffer(jsonBody)

		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateBatchHandler(w, req)
	}
}

// TestProfileMemory выполняет профилирование памяти и сохраняет результат
func TestProfileMemory(t *testing.T) {
	profileName := os.Getenv("PROFILE_NAME")
	if profileName == "" {
		t.Skip("PROFILE_NAME не задан, пропуск теста профилирования")
	}

	runtime.GC()

	runProfiledWorkload(1000)

	dir := "."
	if idx := len(profileName) - len("/"+profileName[len(profileName)-len("base.pprof"):]); idx > 0 {
		for i := len(profileName) - 1; i >= 0; i-- {
			if profileName[i] == '/' {
				dir = profileName[:i]
				break
			}
		}
	}
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	f, err := os.Create(profileName)
	if err != nil {
		t.Fatalf("Не удалось создать файл профиля: %v", err)
	}
	defer f.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		t.Fatalf("Не удалось записать профиль: %v", err)
	}

	t.Logf("Профиль памяти сохранён в %s", profileName)
}

// BenchmarkFullWorkflow тестирует полный рабочий процесс
func BenchmarkFullWorkflow(b *testing.B) {
	h := setupHandler()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ctx := context.WithValue(context.Background(), userIDKey, "test-user")

		body := bytes.NewBufferString("https://example.com/test")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", "text/plain")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CreateHandler(w, req)
	}
}

// BenchmarkBatchWorkflow тестирует batch операции
func BenchmarkBatchWorkflow(b *testing.B) {
	benchmarks := []struct {
		name      string
		batchSize int
	}{
		{"batch_10", 10},
		{"batch_50", 50},
		{"batch_100", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			h := setupHandler()

			requests := make([]model.BatchRequest, bm.batchSize)
			for j := 0; j < bm.batchSize; j++ {
				requests[j] = model.BatchRequest{
					CorrelationID: fmt.Sprintf("id-%d", j),
					OriginalURL:   fmt.Sprintf("https://example.com/%d", j),
				}
			}
			jsonBody, _ := json.Marshal(requests)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ctx := context.WithValue(context.Background(), userIDKey, "test-user")
				body := bytes.NewBuffer(jsonBody)

				req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)

				w := httptest.NewRecorder()
				h.CreateBatchHandler(w, req)
			}
		})
	}
}
