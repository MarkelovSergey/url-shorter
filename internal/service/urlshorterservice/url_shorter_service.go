package urlshorterservice

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/MarkelovSergey/url-shorter/internal/model"
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
	GetOriginalURL(ctx context.Context, id string) (string, error)
	Generate(ctx context.Context, url, userID string) (string, error)
	GenerateBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error)
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

func (s *urlShorterService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	url, err := s.urlShorterRepo.Find(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", fmt.Errorf("%w: %s", service.ErrFindShortCode, shortCode)
		}

		return "", fmt.Errorf("%w: %w", service.ErrFindShortCode, err)
	}

	return url, nil
}

func (s *urlShorterService) Generate(ctx context.Context, url, userID string) (string, error) {
	for i := 0; i < maxGenerateAttempts; i++ {
		candidate := s.generateRandomShortCode()
		resultCode, err := s.urlShorterRepo.Add(ctx, candidate, url, userID)
		if err == nil {
			return resultCode, nil
		}

		if errors.Is(err, repository.ErrURLAlreadyExists) {
			return resultCode, fmt.Errorf("%w: %w", service.ErrURLConflict, err)
		}

		if errors.Is(err, repository.ErrShortCodeAlreadyExist) {
			continue
		}

		return "", fmt.Errorf("%w: %w", service.ErrSaveShortCode, err)
	}

	return "",
		fmt.Errorf("%w after %d attempts", service.ErrGenerateShortCode, maxGenerateAttempts)
}

func (s *urlShorterService) GenerateBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
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

	shortCodes, err := s.urlShorterRepo.AddBatch(ctx, urlMap, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", service.ErrSaveShortCode, err)
	}

	return shortCodes, nil
}

func (s *urlShorterService) GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error) {
	return s.urlShorterRepo.GetUserURLs(ctx, userID)
}

func (s *urlShorterService) generateRandomShortCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
