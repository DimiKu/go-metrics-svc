package server

import (
	"context"
	"go-metric-svc/dto"
	"go-metric-svc/internal/customErrors"
	"go-metric-svc/internal/entities/server"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateValue(metricName string, metricValue float64, ctx context.Context)
	SumValue(metricName string, metricValue int64, ctx context.Context) int64
	GetMetricByName(metricName dto.MetricServiceDto, ctx context.Context) (dto.MetricServiceDto, error)
	GetAllMetrics(ctx context.Context) []string
	DBPing(ctx context.Context) (bool, error)
	SaveMetrics(ctx context.Context, metrics dto.MetricCollectionDto) error
}

type MetricCollectorSvc struct {
	storage Storage
	//dbStorage  DBStorage

	log *zap.SugaredLogger
}

func NewMetricCollectorSvc(
	memStorage Storage,
	log *zap.SugaredLogger,
) *MetricCollectorSvc {
	return &MetricCollectorSvc{
		storage: memStorage,
		log:     log,
	}
}

func (s *MetricCollectorSvc) UpdateStorage(metricName string, metricValue float64, ctx context.Context) {
	//s.log.Info("Update in service")

	s.storage.UpdateValue(metricName, metricValue, ctx)
}

func (s *MetricCollectorSvc) SumInStorage(metricName string, metricValue int64, ctx context.Context) int64 {
	s.log.Info("Sum metric in service")
	newValue := s.storage.SumValue(metricName, metricValue, ctx)
	return newValue
}

func (s *MetricCollectorSvc) GetMetricByName(metric dto.MetricServiceDto, ctx context.Context) (dto.MetricServiceDto, error) {
	collectedMetric, err := s.storage.GetMetricByName(metric, ctx)
	if err != nil {
		return dto.MetricServiceDto{}, customerrors.ErrMetricNotExist
	}
	return collectedMetric, nil
}

func (s *MetricCollectorSvc) GetAllMetrics(ctx context.Context) []string {
	return s.storage.GetAllMetrics(ctx)
}

func (s *MetricCollectorSvc) DBPing(ctx context.Context) (bool, error) {
	return s.storage.DBPing(ctx)
}

func (s *MetricCollectorSvc) CollectMetricsArray(ctx context.Context, metrics []dto.MetricServiceDto) error {
	var metricCollection dto.MetricCollectionDto
	gaugeMetrics := make([]dto.MetricServiceDto, 0)
	counterMetrics := make([]dto.MetricServiceDto, 0)

	for _, m := range metrics {
		switch m.MetricType {
		case server.GaugeMetrics:
			gaugeMetrics = append(gaugeMetrics, m)
		case server.CounterMetrics:
			counterMetrics = append(counterMetrics, m)
		}
	}

	metricCollection.GaugeCollection = gaugeMetrics
	metricCollection.CounterCollection = counterMetrics

	err := s.storage.SaveMetrics(ctx, metricCollection)
	if err != nil {
		return err
	}

	return nil
}
