package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/middleware"
)

func (h *handler) DeleteURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if len(shortURLs) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	h.urlShorterService.DeleteURLsAsync(shortURLs, userID)

	w.WriteHeader(http.StatusAccepted)
}
