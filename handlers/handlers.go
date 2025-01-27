package handlers

import (
	"go-metric-svc/dto"
	"go-metric-svc/utils"
	"strconv"

	"go.uber.org/zap"
	"net/http"
	"strings"
)

type Service interface {
	UpdateStorage(metricName string, num float64)
	SumInStorage(metricName string, num int64)
}

func MetricCollectHandler(service Service, logger *zap.Logger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		req := strings.Split(r.URL.String(), "/")

		if len(req) < 5 {
			http.Error(rw, "metric not found", http.StatusNotFound)
			return
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
		if metricType == dto.MetricTypeServiceCounterTypeDto {
			num, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			service.SumInStorage(metricName, num)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		} else if metricType == dto.MetricTypeServiceGaugeTypeDto {
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
