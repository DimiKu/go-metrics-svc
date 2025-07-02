package ipchecker

import (
	"go.uber.org/zap"
	"net"
	"net/http"
)

func AddrCheckMiddleware(allowedCIDR string, log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Infof("Start handling request with real ip")
			_, network, err := net.ParseCIDR(allowedCIDR)
			log.Infof("Check IP address %s", network)
			if err != nil {
				log.Error("invalid CIDR %s: %v", allowedCIDR, err)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				log.Errorf("X-Real-IP header required")
				return
			}
			log.Infof("Check IP address %s", ipStr)

			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "invalid IP address", http.StatusForbidden)
				log.Errorf("invalid IP address %s", ipStr)
				return
			}

			if !network.Contains(ip) {
				http.Error(w, "IP address not allowed", http.StatusForbidden)
				log.Errorf("IP address not allowed")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
