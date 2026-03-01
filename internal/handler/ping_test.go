package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/usecase/urlcase"
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
		"",
		"",
		false,
	)

	tests := []struct {
		name           string
		method         string
		mockSetup      func(*urlcase.MockURLUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful health check",
			method: http.MethodGet,
			mockSetup: func(m *urlcase.MockURLUseCase) {
				m.EXPECT().Ping(context.Background()).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:   "failed health check",
			method: http.MethodGet,
			mockSetup: func(m *urlcase.MockURLUseCase) {
				m.EXPECT().Ping(context.Background()).Return(errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockUseCase := new(urlcase.MockURLUseCase)

			test.mockSetup(mockUseCase)

			req := httptest.NewRequest(test.method, cfg.Server.Address+"/ping", nil)
			w := httptest.NewRecorder()

			h := New(cfg, mockUseCase, logger)
			h.PingHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, w.Body.String())

			mockUseCase.AssertExpectations(t)
		})
	}
}
