package handlers

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/service/server"
	"go-metric-svc/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFailedPingHandler(t *testing.T) {
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
			name:    "Positive failed ping test #1",
			args:    args{statusCode: 500},
			req:     "/ping",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.req, nil)
			w := httptest.NewRecorder()
			ctx := context.Background()
			handlerFunc := StoragePingHandler(collectorService, ctx, logger)

			handlerFunc(w, req)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.args.statusCode, result.StatusCode)
		})
	}
}

func TestPositivePingHandler(t *testing.T) {
	log, _ := zap.NewProduction()
	logger := log.Sugar()
	ctrl := gomock.NewController(t)
	mockCollectorService := NewMockService(ctrl)

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
			name:    "Positive ping test #1",
			args:    args{statusCode: 200},
			req:     "/ping",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.req, nil)
			w := httptest.NewRecorder()
			ctx := context.Background()
			mockCollectorService.EXPECT().DBPing(ctx).Return(true, nil)
			handlerFunc := StoragePingHandler(mockCollectorService, ctx, logger)

			handlerFunc(w, req)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.args.statusCode, result.StatusCode)
		})
	}
}
