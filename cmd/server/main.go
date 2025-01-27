package main

import (
	"github.com/go-chi/chi/v5"
	"go-metric-svc/handlers"
	"go-metric-svc/service"
	"go-metric-svc/storage"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	logger, _ := zap.NewProduction()

	initialStorage := make(map[string]storage.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, logger)
	collectorService := service.NewMetricCollectorSvc(memStorage, logger)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, logger))
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
