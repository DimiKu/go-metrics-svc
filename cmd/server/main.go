package main

import (
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"go-metric-svc/internal/config"
	"go-metric-svc/internal/handlers"
	"go-metric-svc/internal/service"
	"go-metric-svc/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	var cfg config.ServerConfig
	var addr string
	logger, _ := zap.NewProduction()
	log := logger.Sugar()

	err := env.Parse(&cfg)
	if err != nil {
		log.Errorf("Error parse env: %s", err)
	}

	r := chi.NewRouter()
	parseFlags()

	if cfg.Addr != "" {
		addr = cfg.Addr
	} else {
		addr = flagRunAddr
	}

	initialStorage := make(map[string]storage.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, log)
	collectorService := service.NewMetricCollectorSvc(memStorage, log)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, log))
	r.Get("/value/{metricType}/{metricName}", handlers.MetricReceiveHandler(collectorService, log))
	r.Get("/", handlers.MetricReceiveAllMetricsHandler(collectorService, log))
	log.Infof("Server start on %s", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
