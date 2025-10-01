package urlshorterrepository

import (
	"sync"

	"github.com/MarkelovSergey/url-shorter/internal/repository"
)

type URLShorterRepository interface {
	Add(shortCode, url string) (string, error)
	Find(shortCode string) (string, error)
}

type urlShorterRepository struct {
	urls       map[string]string
	shortCodes map[string]string
	mu         *sync.Mutex
}

func New() URLShorterRepository {
	return &urlShorterRepository{
		urls:       make(map[string]string),
		shortCodes: make(map[string]string),
		mu:         &sync.Mutex{},
	}
}

func (r *urlShorterRepository) Add(shortCode, url string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingURL, exists := r.urls[shortCode]; exists {
		if existingURL != url {
			return "", repository.ErrURLAlreadyExists
		}

		return shortCode, nil
	}

	if existingShortCode, exists := r.shortCodes[url]; exists {
		return existingShortCode, nil
	}

	r.urls[shortCode] = url
	r.shortCodes[url] = shortCode

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
