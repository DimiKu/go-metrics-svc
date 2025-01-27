package main

import (
	"github.com/go-chi/chi/v5"
	"go-metric-svc/handlers"
	"go-metric-svc/internal/service"
	"go-metric-svc/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	logger, _ := zap.NewProduction()
	log := logger.Sugar()

	initialStorage := make(map[string]storage.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, log)
	collectorService := service.NewMetricCollectorSvc(memStorage, log)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, log))
	r.Get("/value/{metricType}/{metricName}", handlers.MetricReceiveHandler(collectorService, log))
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
