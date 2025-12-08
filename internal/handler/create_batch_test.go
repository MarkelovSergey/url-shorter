package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/service/healthservice"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCreateBatchHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}

	tests := []struct {
		name               string
		requestBody        interface{}
		contentType        string
		mockSetup          func(*urlshorterservice.MockURLShorterService)
		expectedStatusCode int
		validateResponse   func(*testing.T, []model.BatchResponse)
	}{
		{
			name: "Successful batch creation",
			requestBody: []model.BatchRequest{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
				{CorrelationID: "2", OriginalURL: "https://google.com"},
			},
			contentType: "application/json",
			mockSetup: func(m *urlshorterservice.MockURLShorterService) {
				m.EXPECT().GenerateBatch([]string{"https://example.com", "https://google.com"}).
					Return([]string{"abc12345", "def67890"}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			validateResponse: func(t *testing.T, resp []model.BatchResponse) {
				assert.Len(t, resp, 2)
				assert.Equal(t, "1", resp[0].CorrelationID)
				assert.Equal(t, "http://localhost:8080/abc12345", resp[0].ShortURL)
				assert.Equal(t, "2", resp[1].CorrelationID)
				assert.Equal(t, "http://localhost:8080/def67890", resp[1].ShortURL)
			},
		},
		{
			name:               "Empty batch",
			requestBody:        []model.BatchRequest{},
			contentType:        "application/json",
			mockSetup:          func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Missing correlation_id",
			requestBody: []model.BatchRequest{
				{CorrelationID: "", OriginalURL: "https://example.com"},
			},
			contentType:        "application/json",
			mockSetup:          func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Invalid URL",
			requestBody: []model.BatchRequest{
				{CorrelationID: "1", OriginalURL: "not-a-url"},
			},
			contentType:        "application/json",
			mockSetup:          func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Wrong content type",
			requestBody:        `{"correlation_id":"1","original_url":"https://example.com"}`,
			contentType:        "text/plain",
			mockSetup:          func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Invalid JSON",
			requestBody:        `{invalid json}`,
			contentType:        "application/json",
			mockSetup:          func(m *urlshorterservice.MockURLShorterService) {},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockURLShorterService := new(urlshorterservice.MockURLShorterService)
			mockHealthService := new(healthservice.MockHealthService)

			test.mockSetup(mockURLShorterService)

			h := New(cfg, mockURLShorterService, mockHealthService, logger)

			var body []byte
			var err error
			if str, ok := test.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(test.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
			req.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			h.CreateBatchHandler(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)

			if test.validateResponse != nil && w.Code == http.StatusCreated {
				var resp []model.BatchResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				test.validateResponse(t, resp)
			}

			mockURLShorterService.AssertExpectations(t)
		})
	}
}

func TestCreateBatchHandlerServiceError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}

	mockURLShorterService := new(urlshorterservice.MockURLShorterService)
	mockHealthService := new(healthservice.MockHealthService)

	mockURLShorterService.EXPECT().GenerateBatch(mock.Anything).Return([]string{}, assert.AnError)

	h := New(cfg, mockURLShorterService, mockHealthService, logger)

	requestBody := []model.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateBatchHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
