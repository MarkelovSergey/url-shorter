package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/model"
)

func (h *handler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		
		return
	}

	records, err := h.urlShorterService.GetUserURLs(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user URLs: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if len(records) == 0 {
		w.WriteHeader(http.StatusNoContent)
		
		return
	}

	response := make([]model.UserURLResponse, 0, len(records))
	for _, record := range records {
		shortURL, err := url.JoinPath(h.config.BaseURL, record.ShortURL)
		if err != nil {
			h.logger.Error("Failed to join URL: " + err.Error())

			continue
		}

		response = append(response, model.UserURLResponse{
			ShortURL:    shortURL,
			OriginalURL: record.OriginalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response: " + err.Error())
	}
}
