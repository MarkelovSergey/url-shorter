package urlshorterrepository

import "sync"

type URLShorterRepository struct {
	urls map[string]string
	mu   sync.RWMutex
}

func New() *URLShorterRepository {
	return &URLShorterRepository{
		urls: make(map[string]string),
	}
}

func (r *URLShorterRepository) Add(shortCode string, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.urls[shortCode] = url
}

func (r *URLShorterRepository) Find(shortCode string) *string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.urls[shortCode]
	if !ok {
		return nil
	}

	return &v
}
