package urlshorterrepository

import (
	"context"
	"fmt"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/storage/memorystorage"
)

func BenchmarkRepositoryAdd(b *testing.B) {
	storage := memorystorage.New()
	repo := New(storage)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Add(ctx, fmt.Sprintf("short%d", i), fmt.Sprintf("https://example.com/%d", i), "user1")
	}
}

func BenchmarkRepositoryFind(b *testing.B) {
	storage := memorystorage.New()
	repo := New(storage)
	ctx := context.Background()

	for i := 0; i < 1000; i++ {
		repo.Add(ctx, fmt.Sprintf("short%d", i), fmt.Sprintf("https://example.com/%d", i), "user1")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Find(ctx, "short500")
	}
}

func BenchmarkRepositoryAddBatch(b *testing.B) {
	benchmarks := []struct {
		name      string
		batchSize int
	}{
		{"batch 10", 10},
		{"batch 100", 100},
		{"batch 500", 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			ctx := context.Background()

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				storage := memorystorage.New()
				repo := New(storage)

				urls := make(map[string]string, bm.batchSize)
				for j := 0; j < bm.batchSize; j++ {
					urls[fmt.Sprintf("short%d", j)] = fmt.Sprintf("https://example.com/%d", j)
				}
				b.StartTimer()

				_, _ = repo.AddBatch(ctx, urls, "user1")
			}
		})
	}
}

func BenchmarkRepositoryGetUserURLs(b *testing.B) {
	benchmarks := []struct {
		name      string
		urlCount  int
		userCount int
	}{
		{"100 urls 10 users", 100, 10},
		{"1000 urls 100 users", 1000, 100},
		{"10000 urls 100 users", 10000, 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			storage := memorystorage.New()
			repo := New(storage)
			ctx := context.Background()

			for i := 0; i < bm.urlCount; i++ {
				userID := fmt.Sprintf("user%d", i%bm.userCount)
				repo.Add(ctx, fmt.Sprintf("short%d", i), fmt.Sprintf("https://example.com/%d", i), userID)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = repo.GetUserURLs(ctx, "user1")
			}
		})
	}
}

func BenchmarkRepositoryDeleteBatch(b *testing.B) {
	benchmarks := []struct {
		name       string
		totalCount int
		deleteSize int
	}{
		{"delete 10 from 1000", 1000, 10},
		{"delete 50 from 5000", 5000, 50},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			ctx := context.Background()

			storage := memorystorage.New()
			repo := New(storage)

			shortURLs := make([]string, bm.deleteSize)
			for j := 0; j < bm.totalCount; j++ {
				repo.Add(ctx, fmt.Sprintf("short%d", j), fmt.Sprintf("https://example.com/%d", j), "user1")
				if j < bm.deleteSize {
					shortURLs[j] = fmt.Sprintf("short%d", j)
				}
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = repo.DeleteBatch(ctx, shortURLs, "user1")
			}
		})
	}
}

type mockStorageForBenchmark struct {
	records map[string]model.URLRecord
}

func newMockStorage() *mockStorageForBenchmark {
	return &mockStorageForBenchmark{
		records: make(map[string]model.URLRecord),
	}
}

func (m *mockStorageForBenchmark) Load(ctx context.Context) ([]model.URLRecord, error) {
	result := make([]model.URLRecord, 0, len(m.records))
	for _, r := range m.records {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockStorageForBenchmark) Append(ctx context.Context, record model.URLRecord) error {
	m.records[record.ShortURL] = record
	return nil
}

func (m *mockStorageForBenchmark) AppendBatch(ctx context.Context, records []model.URLRecord) error {
	for _, r := range records {
		m.records[r.ShortURL] = r
	}
	return nil
}

func (m *mockStorageForBenchmark) FindByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	for _, r := range m.records {
		if r.OriginalURL == originalURL {
			return r.ShortURL, nil
		}
	}
	return "", nil
}

func (m *mockStorageForBenchmark) FindByShortURL(ctx context.Context, shortURL string) (string, error) {
	if r, ok := m.records[shortURL]; ok {
		return r.OriginalURL, nil
	}
	return "", nil
}

func (m *mockStorageForBenchmark) FindByUserID(ctx context.Context, userID string) ([]model.URLRecord, error) {
	result := make([]model.URLRecord, 0)
	for _, r := range m.records {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockStorageForBenchmark) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	for _, url := range shortURLs {
		if r, ok := m.records[url]; ok && r.UserID == userID {
			r.IsDeleted = true
			m.records[url] = r
		}
	}
	return nil
}

func BenchmarkRepositoryFindWithMapStorage(b *testing.B) {
	storage := newMockStorage()
	repo := New(storage)
	ctx := context.Background()

	for i := 0; i < 10000; i++ {
		repo.Add(ctx, fmt.Sprintf("short%d", i), fmt.Sprintf("https://example.com/%d", i), "user1")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Find(ctx, "short5000")
	}
}
