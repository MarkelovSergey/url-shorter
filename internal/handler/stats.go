package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"go.uber.org/zap"
)

// StatsHandler возвращает статистику сервиса: количество URL и пользователей.
// Доступ разрешён только для IP-адресов из доверенной подсети (trusted_subnet).
// При пустом trusted_subnet или IP-адресе вне подсети возвращает 403 Forbidden.
func (h *handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	trustedSubnet := h.config.Server.TrustedSubnet

	if trustedSubnet == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		h.logger.Error("invalid trusted_subnet CIDR", zap.String("cidr", trustedSubnet), zap.Error(err))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	clientIP := net.ParseIP(realIP)
	if clientIP == nil || !ipNet.Contains(clientIP) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	urls, users, err := h.urlShorterService.GetStats(r.Context())
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
