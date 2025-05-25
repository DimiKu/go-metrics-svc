package example

import (
	"context"
	"encoding/json"
	"go-metric-svc/internal/handlers"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/service/server"
	"go-metric-svc/internal/storage"
	"go-metric-svc/internal/utils"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
)

func ExampleMetricReceiveAllMetricsHandler() {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	ctx := context.Background()

	initialStorage := make(map[string]models.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, sugar)
	collectorService := server.NewMetricCollectorSvc(memStorage, sugar)

	handler := handlers.MetricReceiveAllMetricsHandler(collectorService, sugar, ctx)

	req, _ := http.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		sugar.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func ExampleMetricCollectHandler() {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	ctx := context.Background()

	initialStorage := make(map[string]models.StorageValue)
	memStorage := storage.NewMemStorage(initialStorage, sugar)
	collectorService := server.NewMetricCollectorSvc(memStorage, sugar)

	handler := handlers.MetricCollectHandler(collectorService, sugar, ctx)
	req, _ := http.NewRequest("GET", "/update/counter/testMetric/10", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		sugar.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response utils.Response
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		sugar.Errorf("failed to decode response: %v", err)
	}
	if !response.Status || response.Message.MetricName != "testMetric" || response.Message.MetricValue != "42" {
		sugar.Errorf("handler returned unexpected response: got %+v", response)
	}
}
