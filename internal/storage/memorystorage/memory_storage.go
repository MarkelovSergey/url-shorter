package memorystorage

import (
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

func (ms *memoryStorage) Load() ([]model.URLRecord, error) {
	return ms.records, nil
}

func (ms *memoryStorage) Append(record model.URLRecord) error {
	ms.records = append(ms.records, record)
	return nil
}

func (ms *memoryStorage) AppendBatch(records []model.URLRecord) error {
	if len(records) == 0 {
		return nil
	}

	ms.records = append(ms.records, records...)
	
	return nil
}
