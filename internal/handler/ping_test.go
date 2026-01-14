package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPingHandler(t *testing.T) {
	logger := zap.NewNop()

	cfg := config.New(
		"http://localhost:8080",
		"http://localhost:8080",
		"/var/lib/url-shorter/short-url-db.json",
		"postgres://postgres:password@host.docker.internal:5432/postgres",
		"",
		"",
	)

	tests := []struct {
		name           string
		method         string
		mockSetup      func(*healthservice.MockHealthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful health check",
			method: http.MethodGet,
			mockSetup: func(m *healthservice.MockHealthService) {
				m.EXPECT().Ping(context.Background()).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:   "failed health check",
			method: http.MethodGet,
			mockSetup: func(m *healthservice.MockHealthService) {
				m.EXPECT().Ping(context.Background()).Return(errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockHealthService := new(healthservice.MockHealthService)
			mockURLShorterService := new(urlshorterservice.MockURLShorterService)

			test.mockSetup(mockHealthService)

			req := httptest.NewRequest(test.method, cfg.ServerAddress+"/ping", nil)
			w := httptest.NewRecorder()

			mockAuditPublisher := audit.NewMockPublisher()
			h := New(cfg, mockURLShorterService, mockHealthService, logger, mockAuditPublisher)
			h.PingHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, w.Body.String())

			mockHealthService.AssertExpectations(t)
		})
	}
}
