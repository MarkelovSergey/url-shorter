package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"
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

	us := h.urlShorterService.Generate(u)
	shortURL, err := url.JoinPath(h.config.BaseURL, us)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid URL format"))

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
