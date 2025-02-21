package main

import (
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"go-metric-svc/internal/config"
	"go-metric-svc/internal/handlers"
	customLog "go-metric-svc/internal/logger"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/service/server"
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

	initialStorage := make(map[string]models.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, log)
	collectorService := server.NewMetricCollectorSvc(memStorage, log)

	r.Use(customLog.LogMiddleware(log))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, log))
	r.Post("/update/", handlers.MetricJSONReceiveHandler(collectorService, log))
	r.Get("/value/{metricType}/{metricName}", handlers.MetricReceiveHandler(collectorService, log))
	r.Get("/", handlers.MetricReceiveAllMetricsHandler(collectorService, log))
	log.Infof("Server start on %s", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
