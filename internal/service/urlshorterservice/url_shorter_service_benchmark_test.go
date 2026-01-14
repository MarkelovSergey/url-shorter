package urlshorterservice

import (
	"context"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/storage/memorystorage"
	"go.uber.org/zap"
)

func BenchmarkGenerateRandomShortCode(b *testing.B) {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := New(repo, nil, logger).(*urlShorterService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.generateRandomShortCode()
	}
}

func BenchmarkGenerate(b *testing.B) {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := New(repo, nil, logger)

	ctx := context.Background()
	userID := "test-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Generate(ctx, "https://example.com/test", userID)
	}
}

func BenchmarkGenerateBatch(b *testing.B) {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := New(repo, nil, logger)

	ctx := context.Background()
	userID := "test-user"

	urls := make([]string, 100)
	for i := 0; i < 100; i++ {
		urls[i] = "https://example.com/test"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateBatch(ctx, urls, userID)
	}
}

func BenchmarkGetOriginalURL(b *testing.B) {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := New(repo, nil, logger)

	ctx := context.Background()
	userID := "test-user"

	// Подготовим данные
	shortCode, _ := service.Generate(ctx, "https://example.com/test", userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetOriginalURL(ctx, shortCode)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := New(repo, nil, logger)

	ctx := context.Background()
	userID := "test-user"

	// Подготовим данные - добавим 1000 URL
	for i := 0; i < 1000; i++ {
		_, _ = service.Generate(ctx, "https://example.com/test", userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetUserURLs(ctx, userID)
	}
}

// Бенчмарк для тестирования поиска в большом наборе данных
func BenchmarkGetOriginalURLWithManyRecords(b *testing.B) {
	logger := zap.NewNop()
	storage := memorystorage.New()
	repo := urlshorterrepository.New(storage)
	service := New(repo, nil, logger)

	ctx := context.Background()
	userID := "test-user"

	// Создаём много записей
	var targetShortCode string
	for i := 0; i < 10000; i++ {
		shortCode, _ := service.Generate(ctx, "https://example.com/test", userID)
		if i == 5000 {
			targetShortCode = shortCode
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetOriginalURL(ctx, targetShortCode)
	}
}

// MockStorage для тестирования только генерации кодов без записи
type mockStorage struct{}

func (m *mockStorage) Load(ctx context.Context) ([]model.URLRecord, error) {
	return nil, nil
}

func (m *mockStorage) Append(ctx context.Context, record model.URLRecord) error {
	return nil
}

func (m *mockStorage) AppendBatch(ctx context.Context, records []model.URLRecord) error {
	return nil
}

func (m *mockStorage) FindByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	return "", nil
}

func (m *mockStorage) FindByShortURL(ctx context.Context, shortURL string) (string, error) {
	return "https://example.com", nil
}

func (m *mockStorage) FindByUserID(ctx context.Context, userID string) ([]model.URLRecord, error) {
	return nil, nil
}

func (m *mockStorage) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	return nil
}

func BenchmarkGenerateRandomShortCodeOnly(b *testing.B) {
	logger := zap.NewNop()
	repo := urlshorterrepository.New(&mockStorage{})
	service := New(repo, nil, logger).(*urlShorterService)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = service.generateRandomShortCode()
	}
}
