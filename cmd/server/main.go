package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-metric-svc/internal/config"
	"go-metric-svc/internal/handlers"
	"go-metric-svc/internal/middlewares/decrypt"
	"go-metric-svc/internal/middlewares/gzipper"
	customLog "go-metric-svc/internal/middlewares/logger"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/service/server"
	"go-metric-svc/internal/storage"
	"go-metric-svc/internal/utils"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	// Флаги, которые можно передать при компиляции
	// пример: go build -ldflags "-X main.buildVersion=1.0"
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	var cfg config.ServerConfig

	config.GetBuildInfo(buildVersion, buildDate, buildCommit)

	logger, _ := zap.NewProduction()
	log := logger.Sugar()

	r := chi.NewRouter()
	parseFlagsToStruct()

	serverConf, err := config.ValidateServerConfig(cfg, flags, log)
	if err != nil {
		log.Fatal("Failed to validate server config", zap.Error(err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	// TODO обсудить. Не понял
	//ctx, finish := context.WithTimeout(ctx, 2)

	defer cancel()

	collectorService, initialStorage, pool, conn := configureCollectorServiceAndStorage(serverConf, ctx, log)

	saveDataInterval, err := strconv.Atoi(serverConf.StorageInterval)
	if err != nil {
		log.Error(err)
	}

	storeTicker := time.NewTicker(time.Duration(saveDataInterval) * time.Second)

	signalChan := make(chan os.Signal, 1)
	idleConnsClosed := make(chan struct{})
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	r.Use(customLog.LogMiddleware(log))
	r.Use(gzipper.GzipMiddleware(log))
	if serverConf.UseCrypto != "" {
		key, err := utils.LoadPrivateKey(serverConf.UseCrypto)
		if err != nil {
			log.Fatalf(err.Error())
		}

		r.Use(decrypt.DecryptMiddleware(key))
	}

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.MetricCollectHandler(collectorService, log, ctx))
	r.Post("/update/", handlers.MetricJSONCollectHandler(collectorService, log, ctx, serverConf.UseHash))
	r.Post("/updates/", handlers.MetricJSONArrayCollectHandler(collectorService, log, ctx, serverConf.UseHash))
	r.Post("/value/", handlers.MetricReceiveJSONHandler(collectorService, log, ctx))
	r.Get("/value/{metricType}/{metricName}", handlers.MetricReceiveHandler(collectorService, log, ctx))
	r.Get("/ping", handlers.StoragePingHandler(collectorService, ctx, log))

	r.Get("/", handlers.MetricReceiveAllMetricsHandler(collectorService, log, ctx))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if initialStorage != nil {
		go func() {
			for {
				<-storeTicker.C
				producer, err := storage.NewProducer(serverConf.FileStoragePath, log)
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
			producer, err := storage.NewProducer(serverConf.FileStoragePath, log)
			if err != nil {
				log.Errorf("Failed to create producer: %s", err)
			}

			if err := producer.Write(initialStorage); err != nil {
				log.Errorf("Failed to write data: %s", err)
			}
			if err := srv.Shutdown(ctx); err != nil {
				log.Infof("Failed in gracefull shutdown")
			}
			close(idleConnsClosed)
		}()
	} else {
		go func() {
			<-signalChan
			log.Infof("Start gracefull shutdown and closed db conn")

			conn.Close(ctx)
			pool.Close()
			//finish()

			if err := srv.Shutdown(ctx); err != nil {
				log.Infof("Failed in gracefull shutdown")
			}

			close(idleConnsClosed)
		}()
	}

	log.Infof("Server start on %s", serverConf.Addr)
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Failed to start server: %s", err)
	}

	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")
}

func configureCollectorServiceAndStorage(
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
	if cfg.ConnString != "" {
		fmt.Println(cfg.ConnString)
		conn, err := pgx.Connect(ctx, cfg.ConnString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			panic(err)
		}

		DBConfig, err := pgxpool.ParseConfig(cfg.ConnString)
		if err != nil {
			log.Fatalf("Unable to parse database URL: %v\n", err)
			panic(err)
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
		if cfg.NeedRestore {
			consumer, err := storage.NewConsumer(cfg.FileStoragePath, log)
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
