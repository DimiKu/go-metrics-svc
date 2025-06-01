package server

import (
	"context"
	"github.com/golang/mock/gomock"
	"go-metric-svc/dto"
	"go.uber.org/zap"
	"reflect"
	"testing"
)

func TestMetricCollectorSvc_GetMetricByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	log, _ := zap.NewProduction()
	logger := log.Sugar()
	storage := NewMockStorage(ctrl)
	s := &MetricCollectorSvc{
		storage: storage,
		log:     logger,
	}

	type fields struct {
		storage Storage
		log     *zap.SugaredLogger
	}
	type args struct {
		metric dto.MetricServiceDto
		ctx    context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    dto.MetricServiceDto
		wantErr bool
	}{
		{
			name: "positive test 1",
			args: args{
				metric: dto.MetricServiceDto{
					Name:       "testMetric",
					MetricType: "",
					Value:      "",
				},
				ctx: nil,
			},
			want: dto.MetricServiceDto{
				Name:       "testMetric",
				MetricType: "",
				Value:      "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage.EXPECT().GetMetricByName(tt.args.metric, tt.args.ctx).AnyTimes().Return(tt.args.metric, nil)

			got, err := s.GetMetricByName(tt.args.metric, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetricByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetricByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricCollectorSvc_SumInStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	log, _ := zap.NewProduction()
	logger := log.Sugar()
	storage := NewMockStorage(ctrl)
	s := &MetricCollectorSvc{
		storage: storage,
		log:     logger,
	}

	type fields struct {
		storage Storage
		log     *zap.SugaredLogger
	}
	type args struct {
		metricName  string
		metricValue int64
		ctx         context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "positive test 1",
			args: args{
				metricName:  "TestMetric",
				metricValue: 10,
				ctx:         nil,
			},
			want:    14,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage.EXPECT().SumValue(tt.args.metricName, tt.args.metricValue, tt.args.ctx).AnyTimes().Return(int64(14), nil)

			got, err := s.SumInStorage(tt.args.metricName, tt.args.metricValue, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetricByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetricByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
