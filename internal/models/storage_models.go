package models

type StorageValue struct {
	Counter int64
	Gauge   float64
}

type DBMetricServiceDto struct {
	Name  string `db:"metric_name"`
	Value string `db:"metric_value"`
}
