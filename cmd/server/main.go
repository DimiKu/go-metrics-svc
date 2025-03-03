package main

import (
	"github.com/go-chi/chi/v5"
	"go-metric-svc/internal/config"
	"go-metric-svc/internal/handlers"
	"go-metric-svc/internal/middlewares/gzipper"
	customLog "go-metric-svc/internal/middlewares/logger"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/service/server"
	"go-metric-svc/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	var cfg config.ServerConfig

	var (
		addr                   string
		saveInterval           string
		filePathToStoreMetrics string
	)

	logger, _ := zap.NewProduction()
	log := logger.Sugar()

	r := chi.NewRouter()
	parseFlags()

	addr, saveInterval, filePathToStoreMetrics = config.ValidateServerConfig(cfg, flagRunAddr, storeInterval, fileStoragePath)
	initialStorage := make(map[string]models.StorageValue)

	if cfg.NeedRestore || needRestore {
		consumer, err := storage.NewConsumer(filePathToStoreMetrics, log)
		if err != nil {
			log.Errorf("Failed to create consumer: %s", err)
		}

		initialStorage, err = consumer.ReadMetrics()
		if err != nil {
			log.Errorf("Failed to load metris: %s", err)
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	memStorage := storage.NewMemStorage(initialStorage, log)
	collectorService := server.NewMetricCollectorSvc(memStorage, log)

	saveDataInterval, err := strconv.Atoi(saveInterval)
	if err != nil {
		log.Error(err)
	}

	storeTicker := time.NewTicker(time.Duration(saveDataInterval) * time.Second)

	go func() {
		for {
			<-storeTicker.C
			producer, err := storage.NewProducer(filePathToStoreMetrics, log)
			if err != nil {
				log.Errorf("Failed to create producer: %s", err)
			}

			if err := producer.Write(initialStorage); err != nil {
				log.Errorf("Failed to write data: %s", err)
			}
		}
	}()

	go func() {
		<-signalChan
		producer, err := storage.NewProducer(filePathToStoreMetrics, log)
		if err != nil {
			log.Errorf("Failed to create producer: %s", err)
		}

		if err := producer.Write(initialStorage); err != nil {
			log.Errorf("Failed to write data: %s", err)
		}
		os.Exit(0)
	}()

	r.Use(customLog.LogMiddleware(log))
	r.Use(gzipper.GzipMiddleware(log))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, log))
	r.Post("/update/", handlers.MetricJSONCollectHandler(collectorService, log))
	r.Post("/value/", handlers.MetricReceiveJSONHandler(collectorService, log))
	r.Get("/value/{metricType}/{metricName}", handlers.MetricReceiveHandler(collectorService, log))

	r.Get("/", handlers.MetricReceiveAllMetricsHandler(collectorService, log))
	log.Infof("Server start on %s", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
