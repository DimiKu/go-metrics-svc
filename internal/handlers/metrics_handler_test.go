package handlers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/service/server"
	"go-metric-svc/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricCollectHandler(t *testing.T) {
	log, _ := zap.NewProduction()
	logger := log.Sugar()
	initialStorage := make(map[string]models.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, logger)
	collectorService := server.NewMetricCollectorSvc(memStorage, logger)

	type args struct {
		statusCode int
	}
	tests := []struct {
		name    string
		args    args
		req     string
		wantErr bool
	}{
		{
			name: "positive test1",
			args: args{
				statusCode: http.StatusOK,
			},
			wantErr: false,
			req:     "/update/gauge/GCCPUFraction/0.000000",
		},
		{
			name: "positive test2",
			args: args{
				statusCode: http.StatusNotFound,
			},
			wantErr: false,
			req:     "/update/GCCPUFraction/0.000000",
		},
		{
			name: "positive test3",
			args: args{
				statusCode: http.StatusNotFound,
			},
			wantErr: false,
			req:     "/update/GCCPUFraction/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.req, nil)
			w := httptest.NewRecorder()
			ctx := context.Background()
			handlerFunc := MetricCollectHandler(collectorService, logger, ctx)
			handlerFunc(w, req)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.args.statusCode, result.StatusCode)
		})
	}
}

func TestMetricReceiveHandler(t *testing.T) {
	log, _ := zap.NewProduction()
	logger := log.Sugar()

	initialStorage := make(map[string]models.StorageValue)
	initialStorage["GCCPUFraction"] = models.StorageValue{
		Gauge: 0.000000,
	}
	memStorage := storage.NewMemStorage(initialStorage, logger)
	collectorService := server.NewMetricCollectorSvc(memStorage, logger)

	type args struct {
		statusCode int
	}
	tests := []struct {
		name    string
		args    args
		req     string
		wantErr bool
	}{
		{
			name: "positive test1",
			args: args{
				statusCode: http.StatusOK,
			},
			wantErr: false,
			req:     "/update/gauge/GCCPUFraction/0.000000",
		},
		{
			name: "positive test2",
			args: args{
				statusCode: http.StatusNotFound,
			},
			wantErr: false,
			req:     "/value/SecretMetricType/test",
		},
		{
			name: "positive test3",
			args: args{
				statusCode: http.StatusNotFound,
			},
			wantErr: false,
			req:     "/value/gauge/SecretMetric",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.req, nil)
			ctx := context.Background()
			w := httptest.NewRecorder()
			handlerFunc := MetricReceiveHandler(collectorService, logger, ctx)
			handlerFunc(w, req)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.args.statusCode, result.StatusCode)
		})
	}
}
