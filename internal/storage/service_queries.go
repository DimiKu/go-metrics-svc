package storage

var (
	PingQuery                = "Select 1"
	CreateGuageMetricTable   = `CREATE TABLE IF NOT EXISTS gauge_metrics (metric_name varchar(255) NOT NULL, metric_value double precision NOT NULL);`
	CreateCounterMetricTable = `CREATE TABLE IF NOT EXISTS counter_metrics (metric_name varchar(255) NOT NULL, metric_value BIGINT NOT NULL);`
)
