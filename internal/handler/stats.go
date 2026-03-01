package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"go.uber.org/zap"
)

// StatsHandler возвращает статистику сервиса: количество URL и пользователей.
// Проверка доступа по trusted_subnet выполняется middleware.TrustedSubnet.
func (h *handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	urls, users, err := h.urlUseCase.GetStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := model.StatsResponse{
		URLs:  urls,
		Users: users,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode stats response", zap.Error(err))
	}
}
