// Package handler содержит HTTP-обработчики для API сокращения URL.
package handler

import (
	"io"
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/usecase/urlcase"
)

// CreateHandler обрабатывает запрос на создание короткой ссылки в формате text/plain.
func (h *handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
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

	u := string(body)

	if !urlcase.IsValidURL(u) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("url not correct"))

		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	result, err := h.urlUseCase.Shorten(r.Context(), u, userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))

		return
	}

	if result.IsConflict {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(result.ShortURL))

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result.ShortURL))
}
