package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipCompressResponse(t *testing.T) {
	tests := []struct {
		name                string
		acceptEncoding      string
		contentType         string
		responseBody        string
		shouldCompress      bool
		expectedContentType string
	}{
		{
			name:                "JSON with gzip support",
			acceptEncoding:      "gzip",
			contentType:         "application/json",
			responseBody:        `{"url":"https://practicum.yandex.ru","result":"http://localhost:8080/abc123"}`,
			shouldCompress:      true,
			expectedContentType: "application/json",
		},
		{
			name:                "HTML with gzip support",
			acceptEncoding:      "gzip",
			contentType:         "text/html",
			responseBody:        "<html><body>Hello World</body></html>",
			shouldCompress:      true,
			expectedContentType: "text/html",
		},
		{
			name:                "JSON without gzip support",
			acceptEncoding:      "",
			contentType:         "application/json",
			responseBody:        `{"url":"https://practicum.yandex.ru"}`,
			shouldCompress:      false,
			expectedContentType: "application/json",
		},
		{
			name:                "Plain text should not compress",
			acceptEncoding:      "gzip",
			contentType:         "text/plain",
			responseBody:        "Hello World",
			shouldCompress:      false,
			expectedContentType: "text/plain",
		},
		{
			name:                "Image should not compress",
			acceptEncoding:      "gzip",
			contentType:         "image/png",
			responseBody:        "binary data",
			shouldCompress:      false,
			expectedContentType: "image/png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			})

			middleware := Gzipping(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rec := httptest.NewRecorder()
			middleware.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			if tt.shouldCompress {
				if resp.Header.Get("Content-Encoding") != "gzip" {
					t.Errorf("Expected Content-Encoding: gzip, got: %s", resp.Header.Get("Content-Encoding"))
				}

				gzReader, err := gzip.NewReader(resp.Body)
				if err != nil {
					t.Fatalf("Failed to create gzip reader: %v", err)
				}
				defer gzReader.Close()

				body, err := io.ReadAll(gzReader)
				if err != nil {
					t.Fatalf("Failed to read gzipped body: %v", err)
				}

				if string(body) != tt.responseBody {
					t.Errorf("Expected body: %s, got: %s", tt.responseBody, string(body))
				}
			} else {
				if resp.Header.Get("Content-Encoding") == "gzip" {
					t.Errorf("Expected no gzip compression, but got Content-Encoding: gzip")
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read body: %v", err)
				}

				if string(body) != tt.responseBody {
					t.Errorf("Expected body: %s, got: %s", tt.responseBody, string(body))
				}
			}

			if !strings.Contains(resp.Header.Get("Content-Type"), tt.expectedContentType) {
				t.Errorf("Expected Content-Type to contain: %s, got: %s", tt.expectedContentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestGzipDecompressRequest(t *testing.T) {
	expectedBody := `{"url":"https://practicum.yandex.ru"}`

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	gzWriter.Write([]byte(expectedBody))
	gzWriter.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		if string(body) != expectedBody {
			t.Errorf("Expected body: %s, got: %s", expectedBody, string(body))
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := Gzipping(handler)

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status: %d, got: %d", http.StatusOK, rec.Code)
	}
}

func TestGzipDecompressAndCompressRequest(t *testing.T) {
	requestBody := `{"url":"https://practicum.yandex.ru"}`
	responseBody := `{"result":"http://localhost:8080/abc123"}`

	var reqBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&reqBuf)
	gzWriter.Write([]byte(requestBody))
	gzWriter.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		if string(body) != requestBody {
			t.Errorf("Expected request body: %s, got: %s", requestBody, string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	})

	middleware := Gzipping(handler)

	req := httptest.NewRequest(http.MethodPost, "/", &reqBuf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got: %s", resp.Header.Get("Content-Encoding"))
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	body, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("Failed to read gzipped response body: %v", err)
	}

	if string(body) != responseBody {
		t.Errorf("Expected response body: %s, got: %s", responseBody, string(body))
	}
}

func TestGzipInvalidGzipRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Gzipping(handler)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid gzip data"))
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status: %d, got: %d", http.StatusInternalServerError, rec.Code)
	}
}
