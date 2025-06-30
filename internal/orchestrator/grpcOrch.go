package orchestrator

import (
	"context"
	"go-metric-svc/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	UpdateStorage(metricName string, metricValue float64, ctx context.Context) error
	SumInStorage(metricName string, metricValue int64, ctx context.Context) (int64, error)
}
type Orchestrator struct {
	service Service
	proto.UnimplementedMetricsServiceServer
	l *zap.SugaredLogger
}

func NewOrchestrator(service Service, logger *zap.SugaredLogger) *Orchestrator {
	return &Orchestrator{service: service, l: logger}
}

func (o *Orchestrator) UpdateMetric(ctx context.Context, req *proto.MetricRequest) (*proto.MetricResponse, error) {
	switch v := req.Value.(type) {
	case *proto.MetricRequest_Delta:
		delta := v.Delta
		o.l.Infof("Received COUNTER metric. ID: %s, Delta: %d", req.Id, delta)
		_, err := o.service.SumInStorage(req.Id, delta, ctx)
		if err != nil {
			return nil, err
		}

	case *proto.MetricRequest_Val:
		val := v.Val
		o.l.Infof("Received GAUGE metric. ID: %s, Value: %f", req.Id, val)
		err := o.service.UpdateStorage(req.Id, val, ctx)
		if err != nil {
			return nil, err
		}

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown metric type")
	}

	return &proto.MetricResponse{Success: true}, nil
}
