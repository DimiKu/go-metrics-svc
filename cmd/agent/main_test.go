package main

import (
	"go-metric-svc/entities/server"
	"go-metric-svc/handlers"
	"go-metric-svc/service"
	"go-metric-svc/storage"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_sendMetrics(t *testing.T) {
	type args struct {
		metricsMap map[string]float32
		log        *zap.SugaredLogger
	}

	logger, _ := zap.NewProduction()
	log := logger.Sugar()

	initialStorage := make(map[string]server.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, logger)
	collectorService := service.NewMetricCollectorSvc(memStorage, logger)
	handler := http.HandlerFunc(handlers.MetricCollectHandler(collectorService, logger))
	srv := httptest.NewServer(handler)
	defer srv.Close()

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendMetrics(tt.args.metricsMap, tt.args.log); (err != nil) != tt.wantErr {
				t.Errorf("sendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
