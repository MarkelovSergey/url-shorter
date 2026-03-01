// Package urlcase содержит сценарии использования (use cases) для работы с короткими URL.
// Слой находится между транспортным уровнем (HTTP, gRPC) и сервисным уровнем,
// устраняя дублирование логики построения коротких URL и публикации аудита.
package urlcase

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/service"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
)

// ShortenResult результат сокращения одного URL.
type ShortenResult struct {
	// ShortURL полный короткий URL (включает базовый адрес сервера).
	ShortURL string
	// IsConflict true, если URL уже был сокращён ранее.
	IsConflict bool
}

// URLPair пара из короткого и оригинального URL.
type URLPair struct {
	ShortURL    string
	OriginalURL string
}

// BatchItem входной элемент пакетного запроса.
type BatchItem struct {
	OriginalURL   string
	CorrelationID string
}

// BatchResult выходной элемент пакетного ответа.
type BatchResult struct {
	ShortURL      string
	CorrelationID string
}

// URLUseCase определяет интерфейс сценариев использования для сокращения URL.
// Реализация инкапсулирует построение полных коротких URL и публикацию аудит-событий,
// позволяя транспортным уровням (HTTP, gRPC) сосредоточиться только на
// разборе запросов и формировании ответов.
type URLUseCase interface {
	// Shorten сокращает URL. Возвращает ShortenResult с IsConflict=true,
	// если URL уже был сокращён ранее.
	Shorten(ctx context.Context, rawURL, userID string) (ShortenResult, error)

	// ShortenBatch сокращает несколько URL за один вызов.
	ShortenBatch(ctx context.Context, items []BatchItem, userID string) ([]BatchResult, error)

	// Expand возвращает оригинальный URL по короткому коду.
	// Возвращает service.ErrURLDeleted, если URL был удалён.
	Expand(ctx context.Context, id string) (string, error)

	// GetUserURLs возвращает все URLs пользователя с полными короткими адресами.
	GetUserURLs(ctx context.Context, userID string) ([]URLPair, error)

	// DeleteURLsAsync асинхронно удаляет URL.
	DeleteURLsAsync(shortURLs []string, userID string)

	// GetStats возвращает статистику: количество URL и уникальных пользователей.
	GetStats(ctx context.Context) (urls int, users int, err error)

	// Ping проверяет доступность базы данных.
	Ping(ctx context.Context) error
}

// IsValidURL проверяет, является ли rawURL корректным URL с протоколом http/https.
// Используется транспортными уровнями для валидации входных данных до вызова use case.
func IsValidURL(rawURL string) bool {
	uParsed, err := url.Parse(rawURL)
	return err == nil && uParsed != nil &&
		(strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://"))
}

// urlUseCase реализует URLUseCase.
type urlUseCase struct {
	baseURL           string
	urlShorterService urlshorterservice.URLShorterService
	healthService     healthservice.HealthService
	auditPublisher    audit.Publisher
}

// New создаёт новый экземпляр URLUseCase.
func New(
	baseURL string,
	urlShorterService urlshorterservice.URLShorterService,
	healthService healthservice.HealthService,
	auditPublisher audit.Publisher,
) URLUseCase {
	return &urlUseCase{
		baseURL:           baseURL,
		urlShorterService: urlShorterService,
		healthService:     healthService,
		auditPublisher:    auditPublisher,
	}
}

// buildShortURL строит полный короткий URL из короткого кода.
func (u *urlUseCase) buildShortURL(shortCode string) (string, error) {
	return url.JoinPath(u.baseURL, shortCode)
}

// Shorten сокращает URL, строит полный короткий адрес и публикует аудит-событие.
func (u *urlUseCase) Shorten(ctx context.Context, rawURL, userID string) (ShortenResult, error) {
	shortCode, err := u.urlShorterService.Generate(ctx, rawURL, userID)

	shortURL, joinErr := u.buildShortURL(shortCode)
	if joinErr != nil {
		return ShortenResult{}, joinErr
	}

	if err != nil {
		if isConflict(err) {
			u.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, rawURL, &userID))
			return ShortenResult{ShortURL: shortURL, IsConflict: true}, nil
		}

		return ShortenResult{}, err
	}

	u.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, rawURL, &userID))

	return ShortenResult{ShortURL: shortURL}, nil
}

// ShortenBatch сокращает несколько URL за один вызов и строит полные короткие адреса.
func (u *urlUseCase) ShortenBatch(ctx context.Context, items []BatchItem, userID string) ([]BatchResult, error) {
	urls := make([]string, len(items))
	for i, item := range items {
		urls[i] = item.OriginalURL
	}

	shortCodes, err := u.urlShorterService.GenerateBatch(ctx, urls, userID)
	if err != nil {
		return nil, err
	}

	results := make([]BatchResult, 0, len(shortCodes))
	for i, shortCode := range shortCodes {
		shortURL, joinErr := u.buildShortURL(shortCode)
		if joinErr != nil {
			continue
		}

		var correlationID string
		if i < len(items) {
			correlationID = items[i].CorrelationID
		}

		results = append(results, BatchResult{
			ShortURL:      shortURL,
			CorrelationID: correlationID,
		})
	}

	return results, nil
}

// Expand возвращает оригинальный URL по короткому коду и публикует аудит-событие.
func (u *urlUseCase) Expand(ctx context.Context, id string) (string, error) {
	originalURL, err := u.urlShorterService.GetOriginalURL(ctx, id)
	if err != nil {
		return "", err
	}

	u.auditPublisher.Publish(audit.NewEvent(audit.ActionFollow, originalURL, nil))

	return originalURL, nil
}

// GetUserURLs возвращает все URL пользователя с построенными полными короткими адресами.
func (u *urlUseCase) GetUserURLs(ctx context.Context, userID string) ([]URLPair, error) {
	records, err := u.urlShorterService.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	pairs := make([]URLPair, 0, len(records))
	for _, record := range records {
		shortURL, joinErr := u.buildShortURL(record.ShortURL)
		if joinErr != nil {
			continue
		}

		pairs = append(pairs, URLPair{
			ShortURL:    shortURL,
			OriginalURL: record.OriginalURL,
		})
	}

	return pairs, nil
}

// DeleteURLsAsync делегирует асинхронное удаление URL сервисному слою.
func (u *urlUseCase) DeleteURLsAsync(shortURLs []string, userID string) {
	u.urlShorterService.DeleteURLsAsync(shortURLs, userID)
}

// GetStats делегирует получение статистики сервисному слою.
func (u *urlUseCase) GetStats(ctx context.Context) (int, int, error) {
	return u.urlShorterService.GetStats(ctx)
}

// Ping проверяет доступность базы данных.
func (u *urlUseCase) Ping(ctx context.Context) error {
	return u.healthService.Ping(ctx)
}

func isConflict(err error) bool {
	return errors.Is(err, service.ErrURLConflict)
}
