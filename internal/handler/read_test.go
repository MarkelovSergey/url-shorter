package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/service"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestReadHandler(t *testing.T) {
	logger := zap.NewNop()

	cfg := config.New(
		"http://localhost:8080",
		"http://localhost:8080",
		"/var/lib/url-shorter/short-url-db.json",
	)

	originalURL := "https://practicum.yandex.ru"
	shortID := "test"

	tests := []struct {
		name           string
		method         string
		path           string
		mockSetup      func(*urlshorterservice.MockURLShorterService)
		expectedStatus int
		expectedBody   string
		expectedURL    string
	}{
		{
			name:   "successful redirection",
			method: http.MethodGet,
			path:   "/" + shortID,
			mockSetup: func(m *urlshorterservice.MockURLShorterService) {
				m.EXPECT().GetOriginalURL(shortID).Return(originalURL, nil)
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    originalURL,
		},
		{
			name:           "Invalid path format",
			method:         http.MethodGet,
			path:           "/some/invalid/path",
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "ID not found",
		},
		{
			name:   "ID not found",
			method: http.MethodGet,
			path:   "/" + shortID,
			mockSetup: func(m *urlshorterservice.MockURLShorterService) {
				m.EXPECT().GetOriginalURL(shortID).Return("", service.ErrFindShortCode)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "ID not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(urlshorterservice.MockURLShorterService)
			test.mockSetup(mockService)

			req := httptest.NewRequest(test.method, test.path, nil)
			w := httptest.NewRecorder()

			h := New(cfg, mockService, logger)
			h.ReadHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)

			if test.expectedBody != "" {
				assert.Equal(t, test.expectedBody, w.Body.String())
			}

			if test.expectedURL != "" {
				assert.Equal(t, test.expectedURL, w.Header().Get("Location"))
			}

			mockService.AssertExpectations(t)
		})
	}
}
