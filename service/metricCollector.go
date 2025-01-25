package service

import (
	"go.uber.org/zap"
)

type Storage interface {
	UpdateValue(metricName string, metricValue string)
	SumValue(metricName string, metricValue string)
}

type MetricCollectorSvc struct {
	memStorage Storage

	log *zap.Logger
}

func NewMetricCollectorSvc(
	memStorage Storage,
	log *zap.Logger,
) *MetricCollectorSvc {
	return &MetricCollectorSvc{
		memStorage: memStorage,
		log:        log,
	}
}

func (s *MetricCollectorSvc) UpdateStorage(metricName string, metricValue string) {
	s.log.Info("Update in service")
	s.memStorage.UpdateValue(metricName, metricValue)
}

func (s *MetricCollectorSvc) SumInStorage(metricName string, metricValue string) {
	s.log.Info("Sum metric in service")
	s.memStorage.SumValue(metricName, metricValue)
}
