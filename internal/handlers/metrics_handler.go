package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-metric-svc/dto"
	customerrors "go-metric-svc/internal/customErrors"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/utils"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

// Service Интерфейс для сервиса
type Service interface {
	UpdateStorage(metricName string, num float64, ctx context.Context) error
	SumInStorage(metricName string, num int64, ctx context.Context) (int64, error)
	GetMetricByName(metric dto.MetricServiceDto, ctx context.Context) (dto.MetricServiceDto, error)
	GetAllMetrics(ctx context.Context) ([]string, error)
	DBPing(ctx context.Context) (bool, error)
	CollectMetricsArray(ctx context.Context, metrics []dto.MetricServiceDto) error
}

func MetricCollectHandler(service Service, log *zap.SugaredLogger, ctx context.Context) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		req := strings.Split(r.URL.String(), "/")
		if len(req) < 5 {
			http.Error(rw, "metric not found", http.StatusNotFound)
			return
		}

		metricType, metricName, metricValue := req[2], req[3], req[4]
		lowerCaseMetricName := metricName
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
			netValue, err := service.SumInStorage(lowerCaseMetricName, num, ctx)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
			response.Status = true

			response.Message.MetricValue = strconv.FormatInt(netValue, 10)

			utils.MakeResponse(rw, response)
			return
		} else if metricType == dto.MetricTypeHandlerGaugeTypeDto {
			num, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
			}
			log.Infof("Collect counter gauge with name: %s", metricName)
			service.UpdateStorage(lowerCaseMetricName, num, ctx)
			response.Status = true
			utils.MakeResponse(rw, response)
			return
		} else {
			http.Error(rw, "Bad request string", http.StatusBadRequest)
		}
	}
}

func MetricReceiveHandler(service Service, log *zap.SugaredLogger, ctx context.Context) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var MetricDto dto.MetricServiceDto

		req := strings.Split(r.URL.String(), "/")

		if len(req) < 3 {
			http.Error(rw, "metric not found", http.StatusNotFound)
			return
		}

		metricType, metricName := req[2], req[3]

		log.Infof("Got GET req with metricType: %s, metricName: %s", metricType, metricName)
		MetricDto.Name = metricName
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

		metric, err := service.GetMetricByName(MetricDto, ctx)
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

// MetricReceiveJSONHandler Хендлер для отправки метрик в формате Json
func MetricReceiveJSONHandler(service Service, log *zap.SugaredLogger, ctx context.Context) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		var buf bytes.Buffer
		var dtoMetric dto.MetricServiceDto
		var resMetric models.Metrics

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

		dtoMetric.Name = metric.ID
		dtoMetric.MetricType = strings.ToLower(metric.MType)
		m, err := service.GetMetricByName(dtoMetric, ctx)
		if errors.Is(err, customerrors.ErrMetricNotExist) {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		resMetric.ID = m.Name
		switch m.MetricType {
		case dto.MetricTypeHandlerCounterTypeDto:
			intValue, err := strconv.ParseInt(m.Value, 10, 64)
			if err != nil {
				fmt.Println("Ошибка:", err)
				return
			}
			resMetric.Delta = &intValue
		case dto.MetricTypeHandlerGaugeTypeDto:
			float64Number, err := strconv.ParseFloat(m.Value, 64)
			if err != nil {
				fmt.Println("Ошибка:", err)
				return
			}
			resMetric.Value = &float64Number
		}
		resMetric.MType = m.MetricType
		utils.MakeMetricResponse(rw, resMetric)
	}
}

// MetricJSONCollectHandler Хендлер для получения метрик в формате json
func MetricJSONCollectHandler(service Service, log *zap.SugaredLogger, ctx context.Context, useHash string) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if useHash != "" && r.Header.Get("HashSHA256") != "" {
			h := hmac.New(sha256.New, []byte(useHash))
			h.Write(buf.Bytes())
			hashBytes := h.Sum(nil)
			hashString := hex.EncodeToString(hashBytes)
			byteHash := []byte(r.Header.Get("HashSHA256"))
			if !(bytes.Equal(byteHash, []byte(hashString))) {
				http.Error(rw, customerrors.ErrHashMissMatch.Error(), http.StatusBadRequest)
				return
			}
		}

		err = json.Unmarshal(buf.Bytes(), &metric)
		if err != nil {
			log.Infof("error with body %s", buf.Bytes())
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		lowerCaseType := strings.ToLower(metric.MType)
		rw.Header().Set("Content-Type", "application/json")
		log.Infof("Got metric %s", metric.ID)
		switch lowerCaseType {
		case dto.MetricTypeHandlerCounterTypeDto:
			if metric.Delta == nil {
				return
			}
			newValue, err := service.SumInStorage(metric.ID, *metric.Delta, ctx)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
			metric.Delta = &newValue

			utils.MakeMetricResponse(rw, metric)
			return
		case dto.MetricTypeHandlerGaugeTypeDto:
			if metric.Value == nil {
				return
			}
			if err := service.UpdateStorage(metric.ID, *metric.Value, ctx); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			utils.MakeMetricResponse(rw, metric)
			return
		}

	}
}

// MetricJSONArrayCollectHandler Хендлер для отправки массива метрик
func MetricJSONArrayCollectHandler(service Service, log *zap.SugaredLogger, ctx context.Context, useHash string) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		var metrics []models.Metrics

		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		log.Infof("Body: %s", buf)

		err = json.Unmarshal(buf.Bytes(), &metrics)
		if err != nil {
			log.Infof("error with body %s", buf.Bytes())
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if useHash != "" {
			h := hmac.New(sha256.New, []byte(useHash))
			h.Write(buf.Bytes())
			hashBytes := h.Sum(nil)
			hashString := hex.EncodeToString(hashBytes)
			byteHash := []byte(r.Header.Get("HashSHA256"))
			if !(bytes.Equal(byteHash, []byte(hashString))) {
				http.Error(rw, customerrors.ErrHashMissMatch.Error(), http.StatusBadRequest)
				return
			}
		}

		serviceMetrics := make([]dto.MetricServiceDto, len(metrics))
		rw.Header().Set("Content-Type", "application/json")
		for n, metric := range metrics {
			var m dto.MetricServiceDto
			m.Name = metric.ID
			m.MetricType = metric.MType
			switch m.MetricType {
			case dto.MetricTypeHandlerGaugeTypeDto:
				m.Value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			case dto.MetricTypeHandlerCounterTypeDto:
				m.Value = strconv.FormatInt(*metric.Delta, 10)
			}

			serviceMetrics[n] = m
		}

		err = service.CollectMetricsArray(ctx, serviceMetrics)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		utils.MakeMetricsResponse(rw, metrics)
	}
}
