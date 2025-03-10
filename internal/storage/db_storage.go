package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go-metric-svc/dto"
	"go-metric-svc/internal/entities/server"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
)

type DBStorage struct {
	conn *pgx.Conn
	ctx  context.Context

	log *zap.SugaredLogger
}

func NewDBStorage(conn *pgx.Conn, log *zap.SugaredLogger, ctx context.Context) *DBStorage {
	_, err := conn.Exec(ctx, CreateCounterMetricTable)
	if err != nil {
		log.Errorf("Failed to create counter metric table: %s", err)
	}

	_, err = conn.Exec(ctx, CreateGuageMetricTable)
	if err != nil {
		log.Errorf("Failed to create gauge metric table: %s", err)
	}

	return &DBStorage{
		conn: conn,
		ctx:  ctx,
		log:  log,
	}
}

func (d *DBStorage) DBPing(ctx context.Context) (bool, error) {
	var result int
	if d.conn == nil {
		return false, nil
	}
	err := d.conn.QueryRow(ctx, PingQuery).Scan(&result)
	if err != nil {
		return false, err
	}
	if result == 1 {
		return true, nil
	}

	return false, err
}

func (d *DBStorage) UpdateValue(metricName string, metricValue float64) {
	var currentValue float32
	GetMetricByNameWithTable := fmt.Sprintf(GetMetricByName, "gauge_metrics")
	err := d.conn.QueryRow(d.ctx, GetMetricByNameWithTable, metricName).Scan(&currentValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			InsertNewMetricValueWithTableName := fmt.Sprintf(InsertNewMetricValue, "gauge_metrics")
			_, err = d.conn.Exec(d.ctx, InsertNewMetricValueWithTableName, metricName, metricValue)
			if err != nil {
				d.log.Errorf("Error in InsertNewMetricValue: %s", err)
			}
		} else {
			d.log.Errorf("Error in UpdateValue: %s", err)
		}
	}
	UpdateMetricValueWithTable := fmt.Sprintf(UpdateMetricValue, "gauge_metrics")
	_, err = d.conn.Exec(d.ctx, UpdateMetricValueWithTable, metricValue, metricName)
	if err != nil {
		d.log.Errorf("Error in save gauge_metrics: %s", err)
	}
}

func (d *DBStorage) SumValue(metricName string, metricValue int64) int64 {
	var currentValue int64
	GetMetricByNameWithTable := fmt.Sprintf(GetMetricByName, "counter_metrics")
	err := d.conn.QueryRow(d.ctx, GetMetricByNameWithTable, metricName).Scan(&currentValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			insertNewMetricValueWithTableName := fmt.Sprintf(InsertNewMetricValue, "counter_metrics")
			_, err = d.conn.Exec(d.ctx, insertNewMetricValueWithTableName, metricName, metricValue)
			if err != nil {
				d.log.Errorf("Error in InsertNewMetricValueWithTableName: %s", err)
			}
			return metricValue
		} else {
			d.log.Errorf("Error in GetMetricByName: %s, metricName: %s", err, metricName)
		}
	}

	newValue := currentValue + metricValue
	UpdateMetricValueWithTable := fmt.Sprintf(UpdateMetricValue, "counter_metrics")
	_, err = d.conn.Exec(d.ctx, UpdateMetricValueWithTable, newValue, metricName)
	if err != nil {
		d.log.Errorf("Error in UpdateMetricValue: %s, metricName: %s", err, metricName)

	}

	return newValue
}

func (d *DBStorage) GetMetricByName(metric dto.MetricServiceDto) (dto.MetricServiceDto, error) {
	var dbMetric models.DBMetricServiceDto
	var resultMetric dto.MetricServiceDto
	if metric.MetricType == server.GaugeMetrics {
		err := d.conn.QueryRow(d.ctx, GetMetricByName, "gauge_metrics", metric.Name).Scan(&dbMetric)
		if err != nil {
			return dto.MetricServiceDto{}, err
		}
	} else {
		err := d.conn.QueryRow(d.ctx, GetMetricByName, "counter_metrics", metric.Name).Scan(&dbMetric)
		if err != nil {
			return dto.MetricServiceDto{}, err
		}
	}

	resultMetric.MetricType = metric.MetricType
	resultMetric.Name = metric.Name
	resultMetric.Value = dbMetric.Value

	return resultMetric, nil
}

func (d *DBStorage) GetAllMetrics() []string {
	return []string{}
}
