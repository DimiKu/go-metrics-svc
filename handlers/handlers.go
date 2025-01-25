package handlers

import (
	"fmt"
	"go-metric-svc/entities/server"
	"go-metric-svc/service"
	"go-metric-svc/utils"
	"strconv"

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
			Status: true,
			Message: struct {
				MetricName  string `json:"name"`
				MetricValue string `json:"value"`
			}{
				MetricName:  metricName,
				MetricValue: metricValue,
			},
		}
		fmt.Println(metricType)
		if metricType == server.CounterMetrics {
			num, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			service.SumInStorage(metricName, num)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		} else if metricType == server.GaugeMetrics {
			num, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			service.UpdateStorage(metricName, num)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		} else {
			http.Error(rw, "Bad request string", http.StatusBadRequest)
		}
	}
}
