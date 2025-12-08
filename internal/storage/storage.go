package storage

import "github.com/MarkelovSergey/url-shorter/internal/model"

type Storage interface {
	Load() ([]model.URLRecord, error)
	Append(record model.URLRecord) error
	AppendBatch(records []model.URLRecord) error
}
