package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/usecase/urlcase"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// exampleTestSetup содержит общую конфигурацию для примеров.
type exampleTestSetup struct {
	cfg         config.Config
	logger      *zap.Logger
	mockUseCase *urlcase.MockURLUseCase
	handler     *handler
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
		"",
		"",
		false,
	)
	logger := zap.NewNop()

	t := &exampleMockTestingT{}
	mockUseCase := urlcase.NewMockURLUseCase(t)

	h := New(cfg, mockUseCase, logger)

	return &exampleTestSetup{
		cfg:         cfg,
		logger:      logger,
		mockUseCase: mockUseCase,
		handler:     h,
	}
}

// Example_createHandler демонстрирует создание короткой ссылки через text/plain эндпоинт.
//
// POST / с Content-Type: text/plain
// Тело запроса содержит оригинальный URL.
// Возвращает короткую ссылку со статусом 201 Created.
func Example_createHandler() {
	setup := newExampleTestSetup()

	setup.mockUseCase.EXPECT().
		Shorten(mock.Anything, "https://practicum.yandex.ru", mock.Anything).
		Return(urlcase.ShortenResult{ShortURL: "http://localhost:8080/abc123"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru"))
	req.Header.Set("Content-Type", "text/plain")

	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

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

	setup.mockUseCase.EXPECT().
		Shorten(mock.Anything, "https://practicum.yandex.ru", mock.Anything).
		Return(urlcase.ShortenResult{ShortURL: "http://localhost:8080/xyz789"}, nil)

	requestBody := `{"url": "https://practicum.yandex.ru"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

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
func Example_createBatchHandler() {
	setup := newExampleTestSetup()

	items := []urlcase.BatchItem{
		{OriginalURL: "https://example.com", CorrelationID: "id1"},
		{OriginalURL: "https://google.com", CorrelationID: "id2"},
	}
	setup.mockUseCase.EXPECT().
		ShortenBatch(mock.Anything, items, mock.Anything).
		Return([]urlcase.BatchResult{
			{ShortURL: "http://localhost:8080/short1", CorrelationID: "id1"},
			{ShortURL: "http://localhost:8080/short2", CorrelationID: "id2"},
		}, nil)

	requestBody := `[
		{"correlation_id": "id1", "original_url": "https://example.com"},
		{"correlation_id": "id2", "original_url": "https://google.com"}
	]`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

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
func Example_readHandler() {
	setup := newExampleTestSetup()

	setup.mockUseCase.EXPECT().
		Expand(mock.Anything, "abc123").
		Return("https://practicum.yandex.ru", nil)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)

	w := httptest.NewRecorder()
	setup.handler.ReadHandler(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Location:", w.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://practicum.yandex.ru
}

// Example_getUserURLsHandler демонстрирует получение списка URL пользователя.
func Example_getUserURLsHandler() {
	setup := newExampleTestSetup()

	setup.mockUseCase.EXPECT().
		GetUserURLs(mock.Anything, "user-123").
		Return([]urlcase.URLPair{
			{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://practicum.yandex.ru"},
			{ShortURL: "http://localhost:8080/xyz789", OriginalURL: "https://google.com"},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

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
func Example_deleteURLsHandler() {
	setup := newExampleTestSetup()

	setup.mockUseCase.EXPECT().
		DeleteURLsAsync([]string{"abc123", "xyz789"}, "user-123").
		Return()

	requestBody := `["abc123", "xyz789"]`
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	setup.handler.DeleteURLsHandler(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 202
}

// Example_pingHandler демонстрирует проверку доступности базы данных.
func Example_pingHandler() {
	setup := newExampleTestSetup()

	setup.mockUseCase.EXPECT().
		Ping(mock.Anything).
		Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	w := httptest.NewRecorder()
	setup.handler.PingHandler(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 200
}
