package storage

// Запросы для работы с бд
var (
	UpdateMetricValue    = "UPDATE %s SET metric_value = $1 WHERE metric_name = $2;"
	InsertNewMetricValue = `INSERT INTO %s (metric_name, metric_value) VALUES ($1, $2);`
	GetMetricByName      = "SELECT metric_value FROM %s WHERE metric_name = $1;"
)
