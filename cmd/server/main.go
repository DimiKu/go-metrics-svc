package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

	addr, saveInterval, filePathToStoreMetrics, connString = config.ValidateServerConfig(cfg, flagRunAddr, storeInterval, fileStoragePath, connString)

	ctx := context.Background()

	collectorService, initialStorage, pool, conn := configureCollectorServiceAndStorage(connString, needRestore, filePathToStoreMetrics, cfg, ctx, log)

	saveDataInterval, err := strconv.Atoi(saveInterval)
	if err != nil {
		log.Error(err)
	}

	storeTicker := time.NewTicker(time.Duration(saveDataInterval) * time.Second)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	if initialStorage != nil {
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
			log.Infof("Start gracefull shutdown")
			producer, err := storage.NewProducer(filePathToStoreMetrics, log)
			if err != nil {
				log.Errorf("Failed to create producer: %s", err)
			}

			if err := producer.Write(initialStorage); err != nil {
				log.Errorf("Failed to write data: %s", err)
			}

			os.Exit(0)
		}()
	} else {
		go func() {
			<-signalChan
			log.Infof("Start gracefull shutdown and closed db conn")

			conn.Close(ctx)
			pool.Close()

			os.Exit(0)
		}()
	}

	r.Use(customLog.LogMiddleware(log))
	r.Use(gzipper.GzipMiddleware(log))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, log, ctx))
	r.Post("/update/", handlers.MetricJSONCollectHandler(collectorService, log, ctx))
	r.Post("/updates/", handlers.MetricJSONArrayCollectHandler(collectorService, log, ctx))
	r.Post("/value/", handlers.MetricReceiveJSONHandler(collectorService, log, ctx))
	r.Get("/value/{metricType}/{metricName}", handlers.MetricReceiveHandler(collectorService, log, ctx))
	r.Get("/ping", handlers.StoragePingHandler(collectorService, ctx, log))

	r.Get("/", handlers.MetricReceiveAllMetricsHandler(collectorService, log, ctx))
	log.Infof("Server start on %s", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}

func configureCollectorServiceAndStorage(
	connString string,
	needRestore bool,
	filePathToStoreMetrics string,
	cfg config.ServerConfig,
	ctx context.Context,
	log *zap.SugaredLogger,
) (
	*server.MetricCollectorSvc,
	map[string]models.StorageValue,
	*pgxpool.Pool,
	*pgx.Conn,
) {
	var collectorService *server.MetricCollectorSvc
	if connString != "" {
		fmt.Println(connString)
		conn, err := pgx.Connect(ctx, connString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}

		DBConfig, err := pgxpool.ParseConfig(connString)
		if err != nil {
			log.Fatalf("Unable to parse database URL: %v\n", err)
			os.Exit(1)
		}

		DBConfig.MaxConns = 300

		pool, err := pgxpool.NewWithConfig(context.Background(), DBConfig)
		if err != nil {
			log.Fatalf("Unable to create connection pool: %v\n", err)
		}

		dbStorage := storage.NewDBStorage(conn, pool, log)
		collectorService = server.NewMetricCollectorSvc(dbStorage, log)
		return collectorService, nil, pool, conn
	} else {
		initialStorage := make(map[string]models.StorageValue)
		memStorage := storage.NewMemStorage(initialStorage, log)
		collectorService = server.NewMetricCollectorSvc(memStorage, log)
		if cfg.NeedRestore || needRestore {
			consumer, err := storage.NewConsumer(filePathToStoreMetrics, log)
			if err != nil {
				log.Errorf("Failed to create consumer: %s", err)
			}

			initialStorage, err = consumer.ReadMetrics(initialStorage)
			if err != nil {
				log.Errorf("Failed to load metris: %s", err)
			}
		}

		return collectorService, initialStorage, nil, nil
	}
}
