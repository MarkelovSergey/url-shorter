package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateAPIHandler(t *testing.T) {
	logger := zap.NewNop()

	cfg := config.New(
		"http://localhost:8080",
		"http://localhost:8080",
		"/var/lib/url-shorter/short-url-db.json",
	)

	originalURL := "https://practicum.yandex.ru"
	shortID := "test"

	expectedShortURL, err := url.JoinPath(cfg.BaseURL, shortID)
	if err != nil {
		t.Fatalf("Failed to join URL paths: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		mockSetup      func(*urlshorterservice.MockURLShorterService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful URL shortening",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"url":"https://practicum.yandex.ru"}`,
			mockSetup: func(m *urlshorterservice.MockURLShorterService) {
				m.EXPECT().Generate(originalURL).Return(shortID, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   expectedShortURL,
		},
		{
			name:           "Unsupported Content-Type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           originalURL,
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unsupported media type",
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":}`,
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "error parsing JSON",
		},
		{
			name:           "Incorrect URL",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":"not-a-url"}`,
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "url not correct",
		},
		{
			name:           "URL without protocol",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"url":"practicum.yandex.ru"}`,
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "url not correct",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(urlshorterservice.MockURLShorterService)
			test.mockSetup(mockService)

			req := httptest.NewRequest(
				test.method,
				cfg.ServerAddress+"/api/shorten",
				strings.NewReader(test.body),
			)

			req.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			h := New(cfg, mockService, logger)
			h.CreateAPIHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)

			if test.expectedStatus == http.StatusCreated {
				var resp model.Response
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, test.expectedBody, resp.Result)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			} else {
				assert.Equal(t, test.expectedBody, w.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
