package urlshorterservice

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository"
	"github.com/MarkelovSergey/url-shorter/internal/repository/healthrepository"
	"github.com/MarkelovSergey/url-shorter/internal/repository/urlshorterrepository"
	"github.com/MarkelovSergey/url-shorter/internal/service"
	"go.uber.org/zap"
)

const (
	shortCodeLength     = 8
	maxGenerateAttempts = 10
	charset             = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
)

var shortCodePool = sync.Pool{
	New: func() any {
		b := make([]byte, shortCodeLength)

		return &b
	},
}

type URLShorterService interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	Generate(ctx context.Context, url, userID string) (string, error)
	GenerateBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.URLRecord, error)
	DeleteURLsAsync(shortURLs []string, userID string)
}

type urlShorterService struct {
	urlShorterRepo urlshorterrepository.URLShorterRepository
	healthRepo     healthrepository.HealthRepository
	logger         *zap.Logger
	rng            *rand.Rand
	mu             *sync.Mutex
}

func New(
	urlShorterRepo urlshorterrepository.URLShorterRepository,
	healthRepo healthrepository.HealthRepository,
	logger *zap.Logger,
) URLShorterService {
	var seed int64
	if err := binary.Read(cryptorand.Reader, binary.BigEndian, &seed); err != nil {
		seed = time.Now().UnixNano()
	}

	return &urlShorterService{
		urlShorterRepo: urlShorterRepo,
		healthRepo:     healthRepo,
		logger:         logger,
		rng:            rand.New(rand.NewSource(seed)),
		mu:             &sync.Mutex{},
	}
}

func (s *urlShorterService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	url, err := s.urlShorterRepo.Find(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrDeleted) {
			return "", service.ErrURLDeleted
		}
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

	urlMap := make(map[string]string, len(urls))
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

func (s *urlShorterService) DeleteURLsAsync(shortURLs []string, userID string) {
	go s.deleteURLsAsyncWorker(shortURLs, userID)
}

func (s *urlShorterService) deleteURLsAsyncWorker(shortURLs []string, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const batchSize = 10
	var wg sync.WaitGroup

	for i := 0; i < len(shortURLs); i += batchSize {
		end := min(i+batchSize, len(shortURLs))

		batch := shortURLs[i:end]
		wg.Add(1)
		go func(urls []string) {
			defer wg.Done()
			if err := s.urlShorterRepo.DeleteBatch(ctx, urls, userID); err != nil {
				s.logger.Error("Failed to delete URLs batch", zap.Error(err))
			}
		}(batch)
	}

	wg.Wait()
}

func (s *urlShorterService) generateRandomShortCode() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	bufPtr := shortCodePool.Get().(*[]byte)
	b := *bufPtr
	charsetLen := len(charset)
	for i := range b {
		b[i] = charset[s.rng.Intn(charsetLen)]
	}

	result := string(b)
	shortCodePool.Put(bufPtr)

	return result
}
