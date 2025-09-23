package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkelovSergey/url-shorter/config"
	"github.com/stretchr/testify/assert"
)

func TestReadHandler(t *testing.T) {
	config := *config.New("http://localhost:8080", "http://localhost:8080")
	originalURL := "https://practicum.yandex.ru"
	shortID := "test"

	tests := []struct {
		name           string
		method         string
		path           string
		mockSetup      func(*MockURLShorterService)
		expectedStatus int
		expectedBody   string
		expectedURL    string
	}{
		{
			name:   "successful redirection",
			method: http.MethodGet,
			path:   "/" + shortID,
			mockSetup: func(m *MockURLShorterService) {
				m.On(urlShorterServiceGetOriginalURL, shortID).Return(originalURL)
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    originalURL,
		},
		{
			name:           "Invalid path format",
			method:         http.MethodGet,
			path:           "/some/invalid/path",
			mockSetup:      func(m *MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "ID not found",
		},
		{
			name:   "ID not found",
			method: http.MethodGet,
			path:   "/" + shortID,
			mockSetup: func(m *MockURLShorterService) {
				m.On(urlShorterServiceGetOriginalURL, shortID).Return(nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "ID not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(MockURLShorterService)
			test.mockSetup(mockService)

			req := httptest.NewRequest(test.method, test.path, nil)
			w := httptest.NewRecorder()

			h := New(config, mockService)
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
