// Package urlshorterrepository содержит репозиторий для работы с короткими ссылками.
package urlshorterrepository

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
)

// URLShorterRepository определяет интерфейс для работы с сокращенными URL.
type URLShorterRepository interface {
	Add(ctx context.Context, shortCode, url, userID string) (string, error)
	Find(ctx context.Context, shortCode string) (string, error)
	AddBatch(ctx context.Context, urls map[string]string, userID string) ([]string, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error)
	DeleteBatch(ctx context.Context, shortURLs []string, userID string) error
}

type urlShorterRepository struct {
	mu      *sync.Mutex
	storage storage.Storage
	counter int
}

// New создает новый экземпляр URLShorterRepository.
func New(storage storage.Storage) URLShorterRepository {
	repo := &urlShorterRepository{
		mu:      &sync.Mutex{},
		storage: storage,
		counter: 0,
	}

	records, err := storage.Load(context.Background())
	if err == nil {
		for _, record := range records {
			if uuid, err := strconv.Atoi(record.UUID); err == nil && uuid > repo.counter {
				repo.counter = uuid
			}
		}
	}

	return repo
}

// Add добавляет новый URL в хранилище.
func (r *urlShorterRepository) Add(ctx context.Context, shortCode, url, userID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingShortCode, err := r.storage.FindByOriginalURL(ctx, url)
	if err == nil && existingShortCode != "" {
		return existingShortCode, repository.ErrURLAlreadyExists
	}

	r.counter++

	record := model.URLRecord{
		UUID:        strconv.Itoa(r.counter),
		ShortURL:    shortCode,
		OriginalURL: url,
		UserID:      userID,
	}

	if err := r.storage.Append(ctx, record); err != nil {
		r.counter--

		if errors.Is(err, repository.ErrURLAlreadyExists) {
			existingShortCode, findErr := r.storage.FindByOriginalURL(ctx, url)
			if findErr == nil && existingShortCode != "" {
				return existingShortCode, repository.ErrURLAlreadyExists
			}
		}

		return "", err
	}

	return shortCode, nil
}

// Find находит оригинальный URL по короткому коду.
func (r *urlShorterRepository) Find(ctx context.Context, shortCode string) (string, error) {
	originalURL, err := r.storage.FindByShortURL(ctx, shortCode)
	if err != nil {
		return "", err
	}

	if originalURL == "" {
		return "", repository.ErrNotFound
	}

	return originalURL, nil
}

// AddBatch добавляет несколько URL в хранилище пакетно.
func (r *urlShorterRepository) AddBatch(ctx context.Context, urls map[string]string, userID string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(urls) == 0 {
		return []string{}, nil
	}

	shortCodes := make([]string, 0, len(urls))
	records := make([]model.URLRecord, 0, len(urls))

	for shortCode, url := range urls {
		existingShortCode, err := r.storage.FindByOriginalURL(ctx, url)
		if err == nil && existingShortCode != "" {
			shortCodes = append(shortCodes, existingShortCode)
			continue
		}

		r.counter++

		record := model.URLRecord{
			UUID:        strconv.Itoa(r.counter),
			ShortURL:    shortCode,
			OriginalURL: url,
			UserID:      userID,
		}

		records = append(records, record)
		shortCodes = append(shortCodes, shortCode)
	}

	if len(records) > 0 {
		if err := r.storage.AppendBatch(ctx, records); err != nil {
			r.counter -= len(records)
			return nil, err
		}
	}

	return shortCodes, nil
}

// GetUserURLs возвращает все URL пользователя.
func (r *urlShorterRepository) GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	return r.storage.FindByUserID(ctx, userID)
}

// DeleteBatch удаляет несколько URL пакетно.
func (r *urlShorterRepository) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	return r.storage.DeleteBatch(ctx, shortURLs, userID)
}
