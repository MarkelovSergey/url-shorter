package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/model"
)

// GetUserURLsHandler обрабатывает запрос на получение списка URL пользователя.
func (h *handler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	pairs, err := h.urlUseCase.GetUserURLs(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user URLs: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if len(pairs) == 0 {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	response := make([]model.UserURLResponse, 0, len(pairs))
	for _, pair := range pairs {
		response = append(response, model.UserURLResponse{
			ShortURL:    pair.ShortURL,
			OriginalURL: pair.OriginalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response: " + err.Error())
	}
}
