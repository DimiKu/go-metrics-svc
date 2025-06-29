package handlers

import (
	"context"
	"go.uber.org/zap"
	"net/http"
)

// StoragePingHandler oхендлер для проверки соединения с бд
func StoragePingHandler(service Service, ctx context.Context, log *zap.SugaredLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		ping, err := service.DBPing(ctx)
		if err != nil {
			http.Error(rw, "Failed connect to db", http.StatusInternalServerError)
			log.Error("Failed connect to db: %s", err)
			return
		}

		if !ping {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}
