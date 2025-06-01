package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-metric-svc/dto"
	customerrors "go-metric-svc/internal/customErrors"
	"go-metric-svc/internal/entities/server"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/utils"
	"go.uber.org/zap"
	"strconv"
)

// DBStorage Струстура реализующая работу с бд
type DBStorage struct {
	conn *pgx.Conn
	pool *pgxpool.Pool

	log *zap.SugaredLogger
}

func NewDBStorage(conn *pgx.Conn, pool *pgxpool.Pool, log *zap.SugaredLogger) *DBStorage {
	ctx := context.Background()
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
		pool: pool,
		log:  log,
	}
}

func (d *DBStorage) DBPing(ctx context.Context) (bool, error) {
	var result int
	if d.conn == nil {
		return false, nil
	}

	err := d.pool.QueryRow(ctx, PingQuery).Scan(&result)
	if err != nil {
		return false, err
	}
	if result == 1 {
		return true, nil
	}

	return false, err
}

func (d *DBStorage) UpdateValue(metricName string, metricValue float64, ctx context.Context) error {
	var currentValue float32
	GetMetricByNameWithTable := fmt.Sprintf(GetMetricByName, models.GaugeTableName)
	res, err := utils.RetryableQuery(ctx, d.pool, d.log, GetMetricByNameWithTable, metricName)
	if err != nil {
		d.log.Errorf("Error in InsertNewMetricValue: %s", err)
		return err
	}
	if res != nil && !res.Next() {
		InsertNewMetricValueWithTableName := fmt.Sprintf(InsertNewMetricValue, models.GaugeTableName)
		_, err := utils.RetryableExec(ctx, d.pool, InsertNewMetricValueWithTableName, metricName, metricValue)
		if err != nil {
			d.log.Errorf("Error in InsertNewMetricValue: %s", err)
			return err
		}
	} else {
		err = res.Scan(&currentValue)
		if err != nil {
			d.log.Errorf("Error in UpdateValue: %s", err)
			return err
		}
	}

	UpdateMetricValueWithTable := fmt.Sprintf(UpdateMetricValue, models.GaugeTableName)
	_, err = utils.RetryableExec(ctx, d.pool, UpdateMetricValueWithTable, metricValue, metricName)
	if err != nil {
		d.log.Errorf("Error in save gauge_metrics: %s", err)
		return err
	}

	return nil
}

func (d *DBStorage) SumValue(metricName string, metricValue int64, ctx context.Context) (int64, error) {
	var currentValue int64
	GetMetricByNameWithTable := fmt.Sprintf(GetMetricByName, models.CounterTableName)
	res, err := utils.RetryableQuery(ctx, d.pool, d.log, GetMetricByNameWithTable, metricName)
	if err != nil {
		d.log.Errorf("Error in SumValue: %s", err)
	}
	if !res.Next() {
		insertNewMetricValueWithTableName := fmt.Sprintf(InsertNewMetricValue, models.CounterTableName)
		_, err = utils.RetryableExec(ctx, d.pool, insertNewMetricValueWithTableName, metricName, metricValue)
		if err != nil {
			d.log.Errorf("Error in InsertNewMetricValueWithTableName: %s", err)
			return 0, err
		}
		return metricValue, nil
	} else {
		err = res.Scan(&currentValue)
		if err != nil {
			d.log.Errorf("Error in UpdateValue: %s", err)
			return 0, err
		}
	}
	err = res.Scan(&currentValue)
	if err != nil {
		d.log.Errorf("Error in Scan SumValue: %s, metricName: %s", err, metricName)
		return 0, err
	}
	newValue := currentValue + metricValue
	UpdateMetricValueWithTable := fmt.Sprintf(UpdateMetricValue, models.CounterTableName)
	_, err = utils.RetryableExec(ctx, d.pool, UpdateMetricValueWithTable, newValue, metricName)
	if err != nil {
		d.log.Errorf("Error in UpdateMetricValue: %s, metricName: %s", err, metricName)
		return 0, err
	}

	return newValue, nil
}

func (d *DBStorage) GetMetricByName(metric dto.MetricServiceDto, ctx context.Context) (dto.MetricServiceDto, error) {
	var dbMetric models.DBMetricServiceDto
	var resultMetric dto.MetricServiceDto
	if metric.MetricType == server.GaugeMetrics {
		GetMetricByNameWithTable := fmt.Sprintf(GetMetricByName, models.GaugeTableName)
		rows, err := utils.RetryableQuery(ctx, d.pool, d.log, GetMetricByNameWithTable, metric.Name)
		if err != nil {
			return dto.MetricServiceDto{}, customerrors.ErrMetricNotExist
		}
		if !rows.Next() {
			return dto.MetricServiceDto{}, customerrors.ErrMetricNotExist
		}
		err = rows.Scan(&dbMetric.Value)
		if err != nil {
			return dto.MetricServiceDto{}, err

		}
	} else {
		GetMetricByNameWithTable := fmt.Sprintf(GetMetricByName, models.CounterTableName)
		rows, err := utils.RetryableQuery(ctx, d.pool, d.log, GetMetricByNameWithTable, metric.Name)
		if err != nil {
			return dto.MetricServiceDto{}, err
		}
		if !rows.Next() {
			return dto.MetricServiceDto{}, customerrors.ErrMetricNotExist
		}
		err = rows.Scan(&dbMetric.Value)
		if err != nil {
			return dto.MetricServiceDto{}, err
		}

	}

	resultMetric.MetricType = metric.MetricType
	resultMetric.Name = metric.Name
	resultMetric.Value = dbMetric.Value

	return resultMetric, nil
}

func (d *DBStorage) GetAllMetrics(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (d *DBStorage) SaveMetrics(ctx context.Context, metrics dto.MetricCollectionDto) error {
	for _, m := range metrics.CounterCollection {
		value, err := strconv.ParseInt(m.Value, 10, 64) // 10 - основание системы счисления, 64 - битность
		if err != nil {
			return err
		}
		if _, err := d.SumValue(m.Name, value, ctx); err != nil {
			return err
		}
	}

	for _, m := range metrics.GaugeCollection {
		value, err := strconv.ParseFloat(m.Value, 64) // 64 - битность
		if err != nil {
			return err
		}

		if err := d.UpdateValue(m.Name, value, ctx); err != nil {
			return err
		}
	}

	return nil
}
