package handlers

import (
	"errors"
	"go-metric-svc/dto"
	"go-metric-svc/internal/customErrors"
	"go-metric-svc/internal/utils"
	"strconv"

	"go.uber.org/zap"
	"net/http"
	"strings"
)

type Service interface {
	UpdateStorage(metricName string, num float64)
	SumInStorage(metricName string, num int64)
	GetMetricByName(metric dto.MetricServiceDto) (dto.MetricServiceDto, error)
}

func MetricCollectHandler(service Service, log *zap.SugaredLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		req := strings.Split(r.URL.String(), "/")

		if len(req) < 5 {
			http.Error(rw, "metric not found", http.StatusNotFound)
			return
		}

		metricType, metricName, metricValue := req[2], req[3], req[4]
		lowerCaseMetricName := strings.ToLower(metricName)
		log.Infof("Got req with metricType: %s, metricName: %s, metricValue: %s", metricType, metricName, metricValue)
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
		if metricType == dto.MetricTypeHandlerCounterTypeDto {
			num, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			log.Infof("Collect counter mertic with name: %s", metricName)
			service.SumInStorage(lowerCaseMetricName, num)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		} else if metricType == dto.MetricTypeHandlerGaugeTypeDto {
			num, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			log.Infof("Collect counter gauge with name: %s", metricName)
			service.UpdateStorage(lowerCaseMetricName, num)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		} else {
			http.Error(rw, "Bad request string", http.StatusBadRequest)
		}
	}
}

func MetricReceiveHandler(service Service, log *zap.SugaredLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var MetricDto dto.MetricServiceDto

		req := strings.Split(r.URL.String(), "/")

		if len(req) < 3 {
			http.Error(rw, "metric not found", http.StatusNotFound)
			return
		}

		metricType, metricName := req[2], req[3]
		lowerCaseMetricName := strings.ToLower(metricName)

		log.Infof("Got GET req with metricType: %s, metricName: %s", metricType, metricName)
		MetricDto.Name = lowerCaseMetricName
		MetricDto.MetricType = metricType
		response := utils.Response{
			Status: true,
			Message: struct {
				MetricName  string `json:"name"`
				MetricValue string `json:"value"`
			}{
				MetricName:  MetricDto.Name,
				MetricValue: "",
			},
		}

		metric, err := service.GetMetricByName(MetricDto)
		if errors.Is(err, customErrors.ErrMetricNotExist) {
			log.Warnf("Not found metric by name: %s", metricName)
			http.Error(rw, "Metric not found", http.StatusNotFound)
			return
			// TODO кажется так быть не должно. Подумай потом еще
		} else if err != nil {
			log.Warn("Failed to get metric by name: %s", metricName)
			http.Error(rw, "Failed to get metric", http.StatusInternalServerError)
			return
		}

		response.Message.MetricValue = metric.Value
		utils.MakeResponse(rw, response)
	}
}
