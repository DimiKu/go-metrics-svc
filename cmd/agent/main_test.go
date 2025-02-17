package main

import (
	"go-metric-svc/internal/handlers"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/service/agent"
	"go-metric-svc/internal/service/server"
	"go-metric-svc/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

func Test_sendMetrics(t *testing.T) {
	type args struct {
		metricsMap map[string]float32
		log        *zap.SugaredLogger
	}

	logger, _ := zap.NewProduction()
	log := logger.Sugar()

	initialStorage := make(map[string]models.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, log)
	collectorService := server.NewMetricCollectorSvc(memStorage, log)
	handler := http.HandlerFunc(handlers.MetricCollectHandler(collectorService, log))

	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handler),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Errorf("Error starting server: %v", err)
		}
	}()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive test1",
			args: args{
				metricsMap: map[string]float32{
					"Counter": 1.000,
				},
				log: log,
			},
			wantErr: false,
		},
		{
			name: "positive test2",
			args: args{
				metricsMap: map[string]float32{
					"Gauge": 1.134341,
				},
				log: log,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := agent.SendMetrics(tt.args.metricsMap, tt.args.log, "localhost:8080"); (err != nil) != tt.wantErr {
				t.Errorf("sendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
