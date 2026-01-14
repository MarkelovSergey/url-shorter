package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/service"
)

func (h *handler) ReadHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID not found"))

		return
	}

	id := parts[len(parts)-1]
	u, err := h.urlShorterService.GetOriginalURL(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrURLDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID not found"))

		return
	}

	h.auditPublisher.Publish(audit.NewEvent(audit.ActionFollow, u, nil))

	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}
