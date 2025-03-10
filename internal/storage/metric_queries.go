package storage

var (
	UpdateMetricValue    = "UPDATE %s SET metric_value = $1 WHERE metric_name = $2"
	InsertNewMetricValue = `INSERT INTO %s (metric_name, metric_value) VALUES ($1, $2)`
	GetMetricByName      = "SELECT metric_value FROM %s WHERE metric_name = $1"
	//GetAllMetrics        = "SELECT * FROM $1 WHERE metric_name = $2"
	//GetMetricsQuantity   = "SELECT COUNT(*) FROM (SELECT COUNT(*) AS count FROM gauge_metrics UNION ALL SELECT COUNT(*) AS count FROM counter_metrics) AS combined_counts"
)
