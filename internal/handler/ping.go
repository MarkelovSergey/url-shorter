package handler

import (
	"net/http"

	"go.uber.org/zap"
)

// PingHandler проверяет доступность базы данных.
func (h *handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.healthService.Ping(r.Context()); err != nil {
		h.logger.Error("health check failed", zap.Error(err))

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.WriteHeader(http.StatusOK)
}
