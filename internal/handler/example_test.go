package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// exampleTestSetup содержит общую конфигурацию для примеров.
type exampleTestSetup struct {
	cfg                config.Config
	logger             *zap.Logger
	mockURLService     *urlshorterservice.MockURLShorterService
	mockHealthService  *healthservice.MockHealthService
	mockAuditPublisher *audit.MockPublisher
	handler            *handler
}

// exampleMockTestingT реализует интерфейс testing.T для примеров.
type exampleMockTestingT struct{}

func (m *exampleMockTestingT) Errorf(format string, args ...any) {}
func (m *exampleMockTestingT) FailNow()                          {}
func (m *exampleMockTestingT) Cleanup(f func())                  {}
func (m *exampleMockTestingT) Logf(format string, args ...any)   {}

// newExampleTestSetup создает тестовую конфигурацию для примеров.
func newExampleTestSetup() *exampleTestSetup {
	cfg := config.New(
		"http://localhost:8080",
		"http://localhost:8080",
		"/var/lib/url-shorter/short-url-db.json",
		"postgres://postgres:password@localhost:5432/postgres",
		"",
		"",
	)
	logger := zap.NewNop()

	t := &exampleMockTestingT{}
	mockURLService := urlshorterservice.NewMockURLShorterService(t)
	mockHealthService := healthservice.NewMockHealthService(t)
	mockAuditPublisher := audit.NewMockPublisher()

	h := New(cfg, mockURLService, mockHealthService, logger, mockAuditPublisher)

	return &exampleTestSetup{
		cfg:                cfg,
		logger:             logger,
		mockURLService:     mockURLService,
		mockHealthService:  mockHealthService,
		mockAuditPublisher: mockAuditPublisher,
		handler:            h,
	}
}

// Example_createHandler демонстрирует создание короткой ссылки через text/plain эндпоинт.
//
// POST / с Content-Type: text/plain
// Тело запроса содержит оригинальный URL.
// Возвращает короткую ссылку со статусом 201 Created.
func Example_createHandler() {
	setup := newExampleTestSetup()

	// Настраиваем мок для генерации короткой ссылки
	setup.mockURLService.EXPECT().
		Generate(mock.Anything, "https://practicum.yandex.ru", mock.Anything).
		Return("abc123", nil)

	// Создаём запрос с URL в теле
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru"))
	req.Header.Set("Content-Type", "text/plain")

	// Добавляем ID пользователя в контекст
	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

	// Записываем ответ
	w := httptest.NewRecorder()
	setup.handler.CreateHandler(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Body:", w.Body.String())

	// Output:
	// Status: 201
	// Body: http://localhost:8080/abc123
}

// Example_createAPIHandler демонстрирует создание короткой ссылки через JSON API.
//
// POST /api/shorten с Content-Type: application/json
// Тело запроса: {"url": "https://practicum.yandex.ru"}
// Возвращает JSON с короткой ссылкой со статусом 201 Created.
func Example_createAPIHandler() {
	setup := newExampleTestSetup()

	// Настраиваем мок для генерации короткой ссылки
	setup.mockURLService.EXPECT().
		Generate(mock.Anything, "https://practicum.yandex.ru", mock.Anything).
		Return("xyz789", nil)

	// Создаём JSON-запрос
	requestBody := `{"url": "https://practicum.yandex.ru"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем ID пользователя в контекст
	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

	// Записываем ответ
	w := httptest.NewRecorder()
	setup.handler.CreateAPIHandler(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))

	var response model.Response
	json.Unmarshal(w.Body.Bytes(), &response)
	fmt.Println("Result:", response.Result)

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Result: http://localhost:8080/xyz789
}

// Example_createBatchHandler демонстрирует пакетное создание коротких ссылок.
//
// POST /api/shorten/batch с Content-Type: application/json
// Тело запроса: массив объектов с correlation_id и original_url
// Возвращает массив объектов с correlation_id и short_url.
func Example_createBatchHandler() {
	setup := newExampleTestSetup()

	// Настраиваем мок для генерации батча коротких ссылок
	setup.mockURLService.EXPECT().
		GenerateBatch(mock.Anything, []string{"https://example.com", "https://google.com"}, mock.Anything).
		Return([]string{"short1", "short2"}, nil)

	// Создаём батч-запрос
	requestBody := `[
		{"correlation_id": "id1", "original_url": "https://example.com"},
		{"correlation_id": "id2", "original_url": "https://google.com"}
	]`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем ID пользователя в контекст
	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

	// Записываем ответ
	w := httptest.NewRecorder()
	setup.handler.CreateBatchHandler(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))

	var responses []model.BatchResponse
	json.Unmarshal(w.Body.Bytes(), &responses)
	for _, resp := range responses {
		fmt.Printf("CorrelationID: %s, ShortURL: %s\n", resp.CorrelationID, resp.ShortURL)
	}

	// Output:
	// Status: 201
	// Content-Type: application/json
	// CorrelationID: id1, ShortURL: http://localhost:8080/short1
	// CorrelationID: id2, ShortURL: http://localhost:8080/short2
}

// Example_readHandler демонстрирует перенаправление по короткой ссылке.
//
// GET /{shortID}
// Возвращает редирект 307 Temporary Redirect на оригинальный URL.
func Example_readHandler() {
	setup := newExampleTestSetup()

	// Настраиваем мок для получения оригинального URL
	setup.mockURLService.EXPECT().
		GetOriginalURL(mock.Anything, "abc123").
		Return("https://practicum.yandex.ru", nil)

	// Создаём GET-запрос к короткой ссылке
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)

	// Записываем ответ
	w := httptest.NewRecorder()
	setup.handler.ReadHandler(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Location:", w.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://practicum.yandex.ru
}

// Example_getUserURLsHandler демонстрирует получение списка URL пользователя.
//
// GET /api/user/urls
// Возвращает JSON-массив всех URL, созданных пользователем.
func Example_getUserURLsHandler() {
	setup := newExampleTestSetup()

	// Настраиваем мок для получения URL пользователя
	userURLs := []model.URLRecord{
		{ShortURL: "abc123", OriginalURL: "https://practicum.yandex.ru"},
		{ShortURL: "xyz789", OriginalURL: "https://google.com"},
	}
	setup.mockURLService.EXPECT().
		GetUserURLs(mock.Anything, "user-123").
		Return(userURLs, nil)

	// Создаём запрос
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	// Добавляем ID пользователя в контекст
	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

	// Записываем ответ
	w := httptest.NewRecorder()
	setup.handler.GetUserURLsHandler(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))

	var responses []model.UserURLResponse
	json.Unmarshal(w.Body.Bytes(), &responses)
	for _, resp := range responses {
		fmt.Printf("ShortURL: %s, OriginalURL: %s\n", resp.ShortURL, resp.OriginalURL)
	}

	// Output:
	// Status: 200
	// Content-Type: application/json
	// ShortURL: http://localhost:8080/abc123, OriginalURL: https://practicum.yandex.ru
	// ShortURL: http://localhost:8080/xyz789, OriginalURL: https://google.com
}

// Example_deleteURLsHandler демонстрирует удаление URL пользователя.
//
// DELETE /api/user/urls
// Тело запроса: JSON-массив коротких URL для удаления.
// Возвращает статус 202 Accepted - удаление выполняется асинхронно.
func Example_deleteURLsHandler() {
	setup := newExampleTestSetup()

	// Настраиваем мок для асинхронного удаления URL
	setup.mockURLService.EXPECT().
		DeleteURLsAsync([]string{"abc123", "xyz789"}, "user-123").
		Return()

	// Создаём запрос с массивом коротких URL для удаления
	requestBody := `["abc123", "xyz789"]`
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем ID пользователя в контекст
	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

	// Записываем ответ
	w := httptest.NewRecorder()
	setup.handler.DeleteURLsHandler(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 202
}

// Example_pingHandler демонстрирует проверку доступности базы данных.
//
// GET /ping
// Возвращает статус 200 OK если база данных доступна.
func Example_pingHandler() {
	setup := newExampleTestSetup()

	// Настраиваем мок для успешного пинга
	setup.mockHealthService.EXPECT().
		Ping(mock.Anything).
		Return(nil)

	// Создаём запрос
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	// Записываем ответ
	w := httptest.NewRecorder()
	setup.handler.PingHandler(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 200
}
