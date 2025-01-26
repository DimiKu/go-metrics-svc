package handlers

import (
	"github.com/stretchr/testify/assert"
	"go-metric-svc/entities/server"
	"go-metric-svc/service"
	"go-metric-svc/storage"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricCollectHandler(t *testing.T) {
	logger, _ := zap.NewProduction()
	initialStorage := make(map[string]server.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, logger)
	collectorService := service.NewMetricCollectorSvc(memStorage, logger)

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
			handlerFunc := MetricCollectHandler(collectorService, logger)
			handlerFunc(w, req)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.args.statusCode, result.StatusCode)
		})
	}
}
