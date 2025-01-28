package main

import (
	"flag"
	"github.com/go-chi/chi/v5"
	"go-metric-svc/internal/handlers"
	"go-metric-svc/internal/service"
	"go-metric-svc/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

var flagRunAddr string

func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}

func main() {

	r := chi.NewRouter()
	logger, _ := zap.NewProduction()
	log := logger.Sugar()
	parseFlags()

	initialStorage := make(map[string]storage.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, log)
	collectorService := service.NewMetricCollectorSvc(memStorage, log)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, log))
	r.Get("/value/{metricType}/{metricName}", handlers.MetricReceiveHandler(collectorService, log))
	r.Get("/", handlers.MetricReceiveAllMetricsHandler(collectorService, log))
	log.Infof("Server start on %s", flagRunAddr)
	err := http.ListenAndServe(flagRunAddr, r)
	if err != nil {
		panic(err)
	}
}
