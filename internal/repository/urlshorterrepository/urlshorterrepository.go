package urlshorterrepository

import (
	"sync"
)

type URLShorterRepository interface {
	Add(string, string)
	Find(string) *string
}

type urlShorterRepository struct {
	urls map[string]string
	mu   *sync.Mutex
}

func New() URLShorterRepository {
	return &urlShorterRepository{
		urls: make(map[string]string),
		mu:   &sync.Mutex{},
	}
}

func (r *urlShorterRepository) Add(shortCode string, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.urls[shortCode] = url
}

func (r *urlShorterRepository) Find(shortCode string) *string {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.urls[shortCode]
	if !ok {
		return nil
	}

	return &v
}
