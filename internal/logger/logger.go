package logger

import (
	"go.uber.org/zap"
	"net/http"
)

func LogMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//start := time.Now()

			//uri := r.RequestURI
			//method := r.Method
			//
			//duration := time.Since(start)

			//logger.Infof("Uri: %s, METHOD: %s, duration: %s", uri, method, duration)
			next.ServeHTTP(w, r)
		})
	}
}
