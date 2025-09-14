package urlshorterservice

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
)

const shortCodeLength = 8

type URLShorterService interface {
	GetOriginalURL(id string) *string
	Generate(url string) string
}

type urlShorterService struct {
	urlShorterRepo urlshorterrepository.URLShorterRepository
}

func New(urlShorterRepo urlshorterrepository.URLShorterRepository) URLShorterService {
	return &urlShorterService{urlShorterRepo}
}

func (s *urlShorterService) GetOriginalURL(shortCode string) *string {
	return s.urlShorterRepo.Find(shortCode)
}

func (s *urlShorterService) Generate(url string) string {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:shortCodeLength])
	shortCode := encoded[:shortCodeLength]

	u := s.urlShorterRepo.Find(shortCode)
	if u == nil {
		s.urlShorterRepo.Add(shortCode, url)
	}

	return shortCode
}
