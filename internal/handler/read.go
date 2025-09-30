package handler

import (
	"net/http"
	"strings"
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
	u, err := h.urlShorterService.GetOriginalURL(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID not found"))

		return
	}

	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}
