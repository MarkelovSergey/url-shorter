package urlshorterrepository

import (
	"errors"
	"strconv"
	"sync"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
)

type URLShorterRepository interface {
	Add(shortCode, url string) (string, error)
	Find(shortCode string) (string, error)
	AddBatch(urls map[string]string) ([]string, error)
}

type urlShorterRepository struct {
	urls       map[string]string
	shortCodes map[string]string
	mu         *sync.Mutex
	storage    storage.Storage
	counter    int
}

func New(storage storage.Storage) URLShorterRepository {
	repo := &urlShorterRepository{
		urls:       make(map[string]string),
		shortCodes: make(map[string]string),
		mu:         &sync.Mutex{},
		storage:    storage,
		counter:    0,
	}

	records, err := storage.Load()
	if err == nil {
		for _, record := range records {
			repo.urls[record.ShortURL] = record.OriginalURL
			repo.shortCodes[record.OriginalURL] = record.ShortURL

			if uuid, err := strconv.Atoi(record.UUID); err == nil && uuid > repo.counter {
				repo.counter = uuid
			}
		}
	}

	return repo
}

func (r *urlShorterRepository) Add(shortCode, url string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingShortCode, exists := r.shortCodes[url]; exists {
		return existingShortCode, repository.ErrURLAlreadyExists
	}

	r.urls[shortCode] = url
	r.shortCodes[url] = shortCode
	r.counter++

	record := model.URLRecord{
		UUID:        strconv.Itoa(r.counter),
		ShortURL:    shortCode,
		OriginalURL: url,
	}

	if err := r.storage.Append(record); err != nil {
		delete(r.urls, shortCode)
		delete(r.shortCodes, url)
		r.counter--

		if errors.Is(err, repository.ErrURLAlreadyExists) {
			existingShortCode, findErr := r.storage.FindByOriginalURL(url)
			if findErr == nil && existingShortCode != "" {
				r.shortCodes[url] = existingShortCode
				return existingShortCode, repository.ErrURLAlreadyExists
			}
		}

		return "", err
	}

	return shortCode, nil
}

func (r *urlShorterRepository) Find(shortCode string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.urls[shortCode]
	if !ok {
		return "", repository.ErrNotFound
	}

	return v, nil
}

func (r *urlShorterRepository) AddBatch(urls map[string]string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(urls) == 0 {
		return []string{}, nil
	}

	shortCodes := make([]string, 0, len(urls))
	records := make([]model.URLRecord, 0, len(urls))

	for shortCode, url := range urls {
		if existingShortCode, exists := r.shortCodes[url]; exists {
			shortCodes = append(shortCodes, existingShortCode)

			continue
		}

		r.urls[shortCode] = url
		r.shortCodes[url] = shortCode
		r.counter++

		record := model.URLRecord{
			UUID:        strconv.Itoa(r.counter),
			ShortURL:    shortCode,
			OriginalURL: url,
		}

		records = append(records, record)
		shortCodes = append(shortCodes, shortCode)
	}

	if len(records) > 0 {
		if err := r.storage.AppendBatch(records); err != nil {
			for _, record := range records {
				delete(r.urls, record.ShortURL)
				delete(r.shortCodes, record.OriginalURL)
				r.counter--
			}

			return nil, err
		}
	}

	return shortCodes, nil
}
