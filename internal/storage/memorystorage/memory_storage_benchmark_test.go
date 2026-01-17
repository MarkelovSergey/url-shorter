package memorystorage

import (
	"context"
	"fmt"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/model"
)

func BenchmarkMemoryStorageFindByShortURL(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"100 records", 100},
		{"1000 records", 1000},
		{"10000 records", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			storage := New()
			ctx := context.Background()

			// Заполняем storage
			for i := 0; i < bm.count; i++ {
				record := model.URLRecord{
					UUID:        fmt.Sprintf("%d", i),
					ShortURL:    fmt.Sprintf("short%d", i),
					OriginalURL: fmt.Sprintf("https://example.com/%d", i),
					UserID:      "user1",
				}
				storage.Append(ctx, record)
			}

			// Ищем запись в середине
			targetShortURL := fmt.Sprintf("short%d", bm.count/2)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = storage.FindByShortURL(ctx, targetShortURL)
			}
		})
	}
}

func BenchmarkMemoryStorageFindByOriginalURL(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"100 records", 100},
		{"1000 records", 1000},
		{"10000 records", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			storage := New()
			ctx := context.Background()

			// Заполняем storage
			for i := 0; i < bm.count; i++ {
				record := model.URLRecord{
					UUID:        fmt.Sprintf("%d", i),
					ShortURL:    fmt.Sprintf("short%d", i),
					OriginalURL: fmt.Sprintf("https://example.com/%d", i),
					UserID:      "user1",
				}
				storage.Append(ctx, record)
			}

			// Ищем запись в середине
			targetOriginalURL := fmt.Sprintf("https://example.com/%d", bm.count/2)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = storage.FindByOriginalURL(ctx, targetOriginalURL)
			}
		})
	}
}

func BenchmarkMemoryStorageAppend(b *testing.B) {
	storage := New()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		record := model.URLRecord{
			UUID:        fmt.Sprintf("%d", i),
			ShortURL:    fmt.Sprintf("short%d", i),
			OriginalURL: fmt.Sprintf("https://example.com/%d", i),
			UserID:      "user1",
		}
		_ = storage.Append(ctx, record)
	}
}

func BenchmarkMemoryStorageAppendBatch(b *testing.B) {
	benchmarks := []struct {
		name      string
		batchSize int
	}{
		{"batch 10", 10},
		{"batch 100", 100},
		{"batch 1000", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			ctx := context.Background()

			records := make([]model.URLRecord, bm.batchSize)
			for i := 0; i < bm.batchSize; i++ {
				records[i] = model.URLRecord{
					UUID:        fmt.Sprintf("%d", i),
					ShortURL:    fmt.Sprintf("short%d", i),
					OriginalURL: fmt.Sprintf("https://example.com/%d", i),
					UserID:      "user1",
				}
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				newStorage := New()
				_ = newStorage.AppendBatch(ctx, records)
			}
		})
	}
}

func BenchmarkMemoryStorageFindByUserID(b *testing.B) {
	benchmarks := []struct {
		name      string
		count     int
		userCount int
	}{
		{"1000 records 10 users", 1000, 10},
		{"10000 records 100 users", 10000, 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			storage := New()
			ctx := context.Background()

			// Заполняем storage записями разных пользователей
			for i := 0; i < bm.count; i++ {
				record := model.URLRecord{
					UUID:        fmt.Sprintf("%d", i),
					ShortURL:    fmt.Sprintf("short%d", i),
					OriginalURL: fmt.Sprintf("https://example.com/%d", i),
					UserID:      fmt.Sprintf("user%d", i%bm.userCount),
				}
				storage.Append(ctx, record)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = storage.FindByUserID(ctx, "user1")
			}
		})
	}
}

func BenchmarkMemoryStorageDeleteBatch(b *testing.B) {
	benchmarks := []struct {
		name       string
		totalCount int
		deleteSize int
	}{
		{"delete 10 from 1000", 1000, 10},
		{"delete 100 from 10000", 10000, 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			ctx := context.Background()

			// Подготавливаем данные один раз
			storage := New()
			shortURLs := make([]string, bm.deleteSize)
			for j := 0; j < bm.totalCount; j++ {
				record := model.URLRecord{
					UUID:        fmt.Sprintf("%d", j),
					ShortURL:    fmt.Sprintf("short%d", j),
					OriginalURL: fmt.Sprintf("https://example.com/%d", j),
					UserID:      "user1",
				}
				storage.Append(ctx, record)
				if j < bm.deleteSize {
					shortURLs[j] = fmt.Sprintf("short%d", j)
				}
			}

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// DeleteBatch идемпотентен - можно вызывать повторно
				_ = storage.DeleteBatch(ctx, shortURLs, "user1")
			}
		})
	}
}
