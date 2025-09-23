package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateHandler(t *testing.T) {
	originalURL := "https://practicum.yandex.ru"
	shortID := "test"
	host := "localhost:8080"
	expectedShortURL := "http://" + host + "/" + shortID

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		mockSetup      func(*MockURLShorterService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful URL shortening",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        originalURL,
			mockSetup: func(m *MockURLShorterService) {
				m.On(urlShorterServiceGenerate, originalURL).Return(shortID)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   expectedShortURL,
		},
		{
			name:           "Unsupported Content-Type",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           originalURL,
			mockSetup:      func(m *MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unsupported media type",
		},
		{
			name:           "Incorrect URL",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "not-a-url",
			mockSetup:      func(m *MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "url not correct",
		},
		{
			name:           "URL without protocol",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "practicum.yandex.ru",
			mockSetup:      func(m *MockURLShorterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "url not correct",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(MockURLShorterService)
			test.mockSetup(mockService)

			req := httptest.NewRequest(test.method, "http://"+host, strings.NewReader(test.body))
			req.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			h := New(mockService)
			h.CreateHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedBody, w.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}
