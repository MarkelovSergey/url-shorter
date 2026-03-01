package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/usecase/urlcase"
	"go.uber.org/zap"
)

// CreateBatchHandler обрабатывает пакетный запрос на создание коротких ссылок.
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

	items := make([]urlcase.BatchItem, 0, len(requests))
	for _, req := range requests {
		if req.CorrelationID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("correlation_id is required"))

			return
		}

		if !urlcase.IsValidURL(req.OriginalURL) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("url not correct"))

			return
		}

		items = append(items, urlcase.BatchItem{
			OriginalURL:   req.OriginalURL,
			CorrelationID: req.CorrelationID,
		})
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	results, err := h.urlUseCase.ShortenBatch(r.Context(), items, userID)
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

	responses := make([]model.BatchResponse, 0, len(results))
	for _, res := range results {
		responses = append(responses, model.BatchResponse{
			CorrelationID: res.CorrelationID,
			ShortURL:      res.ShortURL,
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
