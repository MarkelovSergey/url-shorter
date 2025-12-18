package storage

import (
	"context"

	"github.com/MarkelovSergey/url-shorter/internal/model"
)

type Storage interface {
	Load(ctx context.Context) ([]model.URLRecord, error)
	Append(ctx context.Context, record model.URLRecord) error
	AppendBatch(ctx context.Context, records []model.URLRecord) error
	FindByOriginalURL(ctx context.Context, originalURL string) (string, error)
	FindByShortURL(ctx context.Context, shortURL string) (string, error)
	FindByUserID(ctx context.Context, userID string) ([]model.URLRecord, error)
}
