package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"go.uber.org/zap"
)

func (h *handler) CreateBatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unsupported media type"))

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error reading request body"))

		return
	}
	defer r.Body.Close()

	var requests []model.BatchRequest
	err = json.Unmarshal(body, &requests)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing JSON"))

		return
	}

	if len(requests) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("empty batch"))

		return
	}

	urls := make([]string, 0, len(requests))
	correlationIDs := make([]string, 0, len(requests))
	for _, req := range requests {
		if req.CorrelationID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("correlation_id is required"))

			return
		}

		uParsed, err := url.Parse(req.OriginalURL)
		if err != nil || uParsed == nil ||
			(!strings.HasPrefix(req.OriginalURL, "http://") && !strings.HasPrefix(req.OriginalURL, "https://")) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("url not correct"))

			return
		}

		urls = append(urls, req.OriginalURL)
		correlationIDs = append(correlationIDs, req.CorrelationID)
	}

	shortCodes, err := h.urlShorterService.GenerateBatch(urls)
	if err != nil {
		h.logger.Error("failed to generate batch short codes",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	responses := make([]model.BatchResponse, 0, len(shortCodes))
	for i, shortCode := range shortCodes {
		shortURL, err := url.JoinPath(h.config.BaseURL, shortCode)
		if err != nil {
			h.logger.Error("failed to join URL path",
				zap.Error(err),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
			
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("invalid URL format"))

			return
		}

		responses = append(responses, model.BatchResponse{
			CorrelationID: correlationIDs[i],
			ShortURL:      shortURL,
		})
	}

	jsonResp, err := json.Marshal(responses)
	if err != nil {
		h.logger.Error("failed to marshal JSON response",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResp)
}
