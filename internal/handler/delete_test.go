package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDeleteURLsHandler(t *testing.T) {
	logger := zap.NewNop()

	cfg := config.New(
		"http://localhost:8080",
		"http://localhost:8080",
		"/var/lib/url-shorter/short-url-db.json",
		"postgres://postgres:password@host.docker.internal:5432/postgres",
	)

	userID := "test-user-123"

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		userID         string
		mockSetup      func(*urlshorterservice.MockURLShorterService)
		expectedStatus int
	}{
		{
			name:        "successful deletion request",
			method:      http.MethodDelete,
			contentType: "application/json",
			body:        `["6qxTVvsy", "RTfd56hn", "Jlfd67ds"]`,
			userID:      userID,
			mockSetup: func(m *urlshorterservice.MockURLShorterService) {
				m.EXPECT().DeleteURLsAsync(mock.Anything, userID).Return().Maybe()
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "missing user ID",
			method:         http.MethodDelete,
			contentType:    "application/json",
			body:           `["6qxTVvsy"]`,
			userID:         "",
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodDelete,
			contentType:    "application/json",
			body:           `{"invalid": "json"}`,
			userID:         userID,
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty array",
			method:         http.MethodDelete,
			contentType:    "application/json",
			body:           `[]`,
			userID:         userID,
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed JSON",
			method:         http.MethodDelete,
			contentType:    "application/json",
			body:           `[not valid json]`,
			userID:         userID,
			mockSetup:      func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "single URL deletion",
			method:      http.MethodDelete,
			contentType: "application/json",
			body:        `["6qxTVvsy"]`,
			userID:      userID,
			mockSetup: func(m *urlshorterservice.MockURLShorterService) {
				m.EXPECT().DeleteURLsAsync(mock.Anything, userID).Return().Maybe()
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "multiple URLs deletion",
			method:      http.MethodDelete,
			contentType: "application/json",
			body:        `["url1", "url2", "url3", "url4", "url5"]`,
			userID:      userID,
			mockSetup: func(m *urlshorterservice.MockURLShorterService) {
				m.EXPECT().DeleteURLsAsync(mock.Anything, userID).Return().Maybe()
			},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(urlshorterservice.MockURLShorterService)
			mockHealthService := new(healthservice.MockHealthService)

			test.mockSetup(mockService)

			req := httptest.NewRequest(test.method, "/api/user/urls", bytes.NewBufferString(test.body))
			req.Header.Set("Content-Type", test.contentType)

			if test.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, test.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			h := New(cfg, mockService, mockHealthService, logger)
			h.DeleteURLsHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)

			if test.expectedStatus == http.StatusAccepted {
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}
