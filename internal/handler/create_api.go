package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/model"
)

func (h *handler) CreateAPIHandler(w http.ResponseWriter, r *http.Request) {
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

	var req model.Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing JSON"))

		return
	}

	uParsed, err := url.Parse(req.URL)
	if err != nil || uParsed == nil ||
		(!strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://")) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("url not correct"))

		return
	}

	us, err := h.urlShorterService.Generate(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))

		return
	}

	shortURL, err := url.JoinPath(h.config.BaseURL, us)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid URL format"))

		return
	}

	resp := model.Response{Result: shortURL}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error generating JSON response"))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResp)
}
