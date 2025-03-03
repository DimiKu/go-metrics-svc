package handlers

import (
	"context"
	"go.uber.org/zap"
	"net/http"
)

func StoragePingHandler(service Service, ctx context.Context, log *zap.SugaredLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		ping, err := service.DbPing(ctx)
		if err != nil {
			http.Error(rw, "Failed connect to db", http.StatusInternalServerError)
			return
		}

		if !ping {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}
