package urlshorterservice

import (
	"crypto/sha256"
	"encoding/base64"
	"math/rand"
	"sync"

	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
)

const (
	shortCodeLength = 8
	charset         = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
)

type URLShorterService interface {
	GetOriginalURL(id string) *string
	Generate(url string) string
}

type urlShorterService struct {
	urlShorterRepo urlshorterrepository.URLShorterRepository
	mu             sync.Mutex
}

func New(urlShorterRepo urlshorterrepository.URLShorterRepository) URLShorterService {
	return &urlShorterService{
		urlShorterRepo: urlShorterRepo,
	}
}

func (s *urlShorterService) GetOriginalURL(shortCode string) *string {
	return s.urlShorterRepo.Find(shortCode)
}

func (s *urlShorterService) Generate(url string) string {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	candidate := encoded[:shortCodeLength]

	s.mu.Lock()
	defer s.mu.Unlock()

	u := s.urlShorterRepo.Find(candidate)
	if u == nil {
		s.urlShorterRepo.Add(candidate, url)
		return candidate
	}

	if *u == url {
		return candidate
	}

	return s.generateRandomShortCode()
}

func (s *urlShorterService) generateRandomShortCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
