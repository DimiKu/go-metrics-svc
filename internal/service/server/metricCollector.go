package server

import (
	"context"
	"go-metric-svc/dto"
	"go-metric-svc/internal/customErrors"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateValue(metricName string, metricValue float64)
	SumValue(metricName string, metricValue int64) int64
	GetMetricByName(metricName dto.MetricServiceDto) (dto.MetricServiceDto, error)
	GetAllMetrics() []string
}

type DBStorage interface {
	DBPing(ctx context.Context) (bool, error)
}

type MetricCollectorSvc struct {
	memStorage Storage
	dbStorage  DBStorage

	log *zap.SugaredLogger
}

func NewMetricCollectorSvc(
	memStorage Storage,
	dbStorage DBStorage,
	log *zap.SugaredLogger,
) *MetricCollectorSvc {
	return &MetricCollectorSvc{
		memStorage: memStorage,
		dbStorage:  dbStorage,
		log:        log,
	}
}

func (s *MetricCollectorSvc) UpdateStorage(metricName string, metricValue float64) {
	//s.log.Info("Update in service")

	s.memStorage.UpdateValue(metricName, metricValue)
}

func (s *MetricCollectorSvc) SumInStorage(metricName string, metricValue int64) int64 {
	s.log.Info("Sum metric in service")
	newValue := s.memStorage.SumValue(metricName, metricValue)
	return newValue
}

func (s *MetricCollectorSvc) GetMetricByName(metric dto.MetricServiceDto) (dto.MetricServiceDto, error) {
	collectedMetric, err := s.memStorage.GetMetricByName(metric)
	if err != nil {
		return dto.MetricServiceDto{}, customerrors.ErrMetricNotExist
	}
	return collectedMetric, nil
}

func (s *MetricCollectorSvc) GetAllMetrics() []string {
	return s.memStorage.GetAllMetrics()
}

func (s *MetricCollectorSvc) DBPing(ctx context.Context) (bool, error) {
	return s.dbStorage.DBPing(ctx)
}
