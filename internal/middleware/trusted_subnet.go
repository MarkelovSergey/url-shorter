package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnet возвращает мидлвар, который разрешает доступ только IP-адресам
// из указанной CIDR-подсети (заголовок X-Real-IP).
// Если подсеть пустая или IP вне диапазона — отвечает 403 Forbidden.
func TrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			_, ipNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
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

			next.ServeHTTP(w, r)
		})
	}
}
