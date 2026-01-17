package memorystorage

import (
	"context"
	"sync"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
)

type memoryStorage struct {
	mu               *sync.RWMutex
	records          []model.URLRecord
	shortURLIndex    map[string]int
	originalURLIndex map[string]int
}

func New() storage.Storage {
	return &memoryStorage{
		mu:               &sync.RWMutex{},
		records:          make([]model.URLRecord, 0),
		shortURLIndex:    make(map[string]int),
		originalURLIndex: make(map[string]int),
	}
}

func (ms *memoryStorage) Load(ctx context.Context) ([]model.URLRecord, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]model.URLRecord, len(ms.records))
	copy(result, ms.records)

	return result, nil
}

func (ms *memoryStorage) Append(ctx context.Context, record model.URLRecord) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	idx := len(ms.records)
	ms.records = append(ms.records, record)
	ms.shortURLIndex[record.ShortURL] = idx
	ms.originalURLIndex[record.OriginalURL] = idx

	return nil
}

func (ms *memoryStorage) AppendBatch(ctx context.Context, records []model.URLRecord) error {
	if len(records) == 0 {
		return nil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	startIdx := len(ms.records)
	ms.records = append(ms.records, records...)

	for i, record := range records {
		idx := startIdx + i
		ms.shortURLIndex[record.ShortURL] = idx
		ms.originalURLIndex[record.OriginalURL] = idx
	}

	return nil
}

func (ms *memoryStorage) FindByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if idx, ok := ms.originalURLIndex[originalURL]; ok {
		return ms.records[idx].ShortURL, nil
	}

	return "", nil
}

func (ms *memoryStorage) FindByShortURL(ctx context.Context, shortURL string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if idx, ok := ms.shortURLIndex[shortURL]; ok {
		record := ms.records[idx]
		if record.IsDeleted {
			return "", repository.ErrDeleted
		}

		return record.OriginalURL, nil
	}

	return "", nil
}

func (ms *memoryStorage) FindByUserID(ctx context.Context, userID string) ([]model.URLRecord, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	count := 0
	for _, record := range ms.records {
		if record.UserID == userID {
			count++
		}
	}

	result := make([]model.URLRecord, 0, count)
	for _, record := range ms.records {
		if record.UserID == userID {
			result = append(result, record)
		}
	}

	return result, nil
}

func (ms *memoryStorage) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, url := range shortURLs {
		if idx, ok := ms.shortURLIndex[url]; ok {
			if ms.records[idx].UserID == userID {
				ms.records[idx].IsDeleted = true
			}
		}
	}

	return nil
}
