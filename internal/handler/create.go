package handler

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/service"
)

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

	uParsed, err := url.Parse(u)
	if err != nil || uParsed == nil ||
		(!strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://")) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("url not correct"))

		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	us, err := h.urlShorterService.Generate(r.Context(), u, userID)

	shortURL, joinErr := url.JoinPath(h.config.BaseURL, us)
	if joinErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid URL format"))

		return
	}

	if err != nil {
		if errors.Is(err, service.ErrURLConflict) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))

			h.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, u, &userID))

			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))

	h.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, u, &userID))
}
