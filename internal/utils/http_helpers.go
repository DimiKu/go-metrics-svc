package utils

import (
	"encoding/json"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type Response struct {
	Status  bool `json:"status"`
	Message struct {
		MetricName  string `json:"name"`
		MetricValue string `json:"value"`
	} `json:"message"`
}

func MakeResponse(w http.ResponseWriter, response Response) {
	// TODO gотом узнаешь нужно раскомментить или удалить
	//jsonRes, err := json.Marshal(response.Message)
	//if err != nil {
	//	log.Fatal("can't decode response", zap.Error(err))
	//}
	w.Write([]byte(response.Message.MetricValue))
}

func MakeMetricResponse(w http.ResponseWriter, metric models.Metrics) {
	jsonRes, err := json.Marshal(metric)
	if err != nil {
		log.Fatal("can't decode response", zap.Error(err))
	}
	w.Write(jsonRes)
}

func MakeMetricsResponse(w http.ResponseWriter, metrics []models.Metrics) {
	jsonRes, err := json.Marshal(metrics)
	if err != nil {
		log.Fatal("can't decode response", zap.Error(err))
	}
	w.Write(jsonRes)
}
