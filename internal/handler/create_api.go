package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/service"
	"go.uber.org/zap"
)

// CreateAPIHandler обрабатывает JSON-запрос на создание короткой ссылки.
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

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	us, err := h.urlShorterService.Generate(r.Context(), req.URL, userID)

	shortURL, joinErr := url.JoinPath(h.config.Server.BaseURL, us)
	if joinErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid URL format"))

		return
	}

	resp := model.Response{Result: shortURL}
	jsonResp, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		h.logger.Error("failed to marshal JSON response",
			zap.Error(marshalErr),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		if errors.Is(err, service.ErrURLConflict) {
			w.WriteHeader(http.StatusConflict)
			w.Write(jsonResp)

			h.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, req.URL, &userID))

			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResp)

	h.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, req.URL, &userID))
}
