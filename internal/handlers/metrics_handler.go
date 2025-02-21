package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-metric-svc/dto"
	customerrors "go-metric-svc/internal/customErrors"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/utils"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

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
		if errors.Is(err, customerrors.ErrMetricNotExist) {
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

func MetricReceiveJSONHandler(service Service, log *zap.SugaredLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		var buf bytes.Buffer
		var dtoMetric dto.MetricServiceDto

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(buf.Bytes(), &metric)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		rw.Header().Set("Content-Type", "application/json")

		dtoMetric.Name = strings.ToLower(metric.ID)
		dtoMetric.MetricType = strings.ToLower(metric.MType)
		resMetric, err := service.GetMetricByName(dtoMetric)
		if errors.Is(err, customerrors.ErrMetricNotExist) {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		utils.MakeMetricResponse(rw, resMetric)
	}
}

func MetricJSONCollectHandler(service Service, log *zap.SugaredLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(buf.Bytes(), &metric)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		response := utils.Response{
			Status: true,
			Message: struct {
				MetricName  string `json:"name"`
				MetricValue string `json:"value"`
			}{
				MetricName:  metric.ID,
				MetricValue: "",
			},
		}

		lowerCaseMetricName := strings.ToLower(metric.ID)
		rw.Header().Set("Content-Type", "application/json")
		switch metric.MType {
		case dto.MetricTypeHandlerCounterTypeDto:
			if metric.Delta == nil {
				return
			}
			service.SumInStorage(lowerCaseMetricName, *metric.Delta)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		case dto.MetricTypeHandlerGaugeTypeDto:
			if metric.Value == nil {
				return
			}
			service.UpdateStorage(lowerCaseMetricName, *metric.Value)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		}

	}
}
