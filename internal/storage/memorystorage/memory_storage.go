package memorystorage

import (
	"context"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
)

type memoryStorage struct {
	records []model.URLRecord
}

func New() storage.Storage {
	return &memoryStorage{
		records: make([]model.URLRecord, 0),
	}
}

func (ms *memoryStorage) Load(ctx context.Context) ([]model.URLRecord, error) {
	return ms.records, nil
}

func (ms *memoryStorage) Append(ctx context.Context, record model.URLRecord) error {
	ms.records = append(ms.records, record)
	return nil
}

func (ms *memoryStorage) AppendBatch(ctx context.Context, records []model.URLRecord) error {
	if len(records) == 0 {
		return nil
	}

	ms.records = append(ms.records, records...)

	return nil
}

func (ms *memoryStorage) FindByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	for _, record := range ms.records {
		if record.OriginalURL == originalURL {
			return record.ShortURL, nil
		}
	}

	return "", nil
}

func (ms *memoryStorage) FindByShortURL(ctx context.Context, shortURL string) (string, error) {
	for _, record := range ms.records {
		if record.ShortURL == shortURL {
			return record.OriginalURL, nil
		}
	}

	return "", nil
}

func (ms *memoryStorage) FindByUserID(ctx context.Context, userID string) ([]model.URLRecord, error) {
	result := make([]model.URLRecord, 0)
	for _, record := range ms.records {
		if record.UserID == userID {
			result = append(result, record)
		}
	}
	
	return result, nil
}
