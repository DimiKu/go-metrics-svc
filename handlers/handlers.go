package handlers

import (
	"encoding/json"
	"go-metric-svc/entities/server"
	"go-metric-svc/service"
	"go-metric-svc/utils"

	"go.uber.org/zap"
	"net/http"
	"strings"
)

func MetricCollectHandler(service *service.MetricCollectorSvc, logger *zap.Logger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		req := strings.Split(r.URL.String(), "/")

		if len(req) < 5 {
			http.Error(rw, "metric not found", http.StatusNotFound)
		}

		metricType, metricName, metricValue := req[2], req[3], req[4]
		response := utils.Response{
			Status: false,
			Message: struct {
				MetricName  string `json:"name"`
				MetricValue string `json:"value"`
			}{
				MetricName:  metricName,
				MetricValue: metricValue,
			},
		}

		jsonRes, err := json.Marshal(response)
		if err != nil {
			logger.Error("can't decode response", zap.Error(err))
		}
		if utils.Contains(server.SumMetrics, metricType) {
			service.SumInStorage(metricName, metricValue)
			utils.MakeResponse(rw, jsonRes)
			return

		} else if utils.Contains(server.UpdateMetrics, metricType) {
			service.UpdateStorage(metricName, metricValue)
			utils.MakeResponse(rw, jsonRes)
			return
		} else {
			http.Error(rw, "Bad request string", http.StatusBadRequest)
		}

		utils.MakeResponse(rw, jsonRes)
	}
}
