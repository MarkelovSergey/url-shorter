package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (h *handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("method not allowed"))

		return
	}

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
	if err != nil || uParsed == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("url not correct"))

		return
	}

	us := h.urlShorterService.Generate(u)
	shortURL := fmt.Sprintf("http://%v/%v", r.Host, us)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
