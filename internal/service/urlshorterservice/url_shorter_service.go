package urlshorterservice

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/MarkelovSergey/url-shorter/internal/repository"
	"github.com/MarkelovSergey/url-shorter/internal/repository/healthrepository"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service"
)

const (
	shortCodeLength     = 8
	maxGenerateAttempts = 10
	charset             = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
)

type URLShorterService interface {
	GetOriginalURL(id string) (string, error)
	Generate(url string) (string, error)
	GenerateBatch(urls []string) ([]string, error)
}

type urlShorterService struct {
	urlShorterRepo urlshorterrepository.URLShorterRepository
	healthRepo     healthrepository.HealthRepository
}

func New(
	urlShorterRepo urlshorterrepository.URLShorterRepository,
	healthRepo healthrepository.HealthRepository,
) URLShorterService {
	return &urlShorterService{
		urlShorterRepo: urlShorterRepo,
		healthRepo:     healthRepo,
	}
}

func (s *urlShorterService) GetOriginalURL(shortCode string) (string, error) {
	url, err := s.urlShorterRepo.Find(shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", fmt.Errorf("%w: %s", service.ErrFindShortCode, shortCode)
		}

		return "", fmt.Errorf("%w: %w", service.ErrFindShortCode, err)
	}

	return url, nil
}

func (s *urlShorterService) Generate(url string) (string, error) {
	for i := 0; i < maxGenerateAttempts; i++ {
		candidate := s.generateRandomShortCode()
		resultCode, err := s.urlShorterRepo.Add(candidate, url)
		if err == nil {
			return resultCode, nil
		}

		if errors.Is(err, repository.ErrURLAlreadyExists) {
			continue
		}

		return "", fmt.Errorf("%w: %w", service.ErrSaveShortCode, err)
	}

	return "",
		fmt.Errorf("%w after %d attempts", service.ErrGenerateShortCode, maxGenerateAttempts)
}

func (s *urlShorterService) GenerateBatch(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return []string{}, nil
	}

	urlMap := make(map[string]string, 0)
	for _, url := range urls {
		for i := 0; i < maxGenerateAttempts; i++ {
			candidate := s.generateRandomShortCode()
			if _, exists := urlMap[candidate]; !exists {
				urlMap[candidate] = url

				break
			}
		}
	}

	shortCodes, err := s.urlShorterRepo.AddBatch(urlMap)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", service.ErrSaveShortCode, err)
	}

	return shortCodes, nil
}

func (s *urlShorterService) generateRandomShortCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
