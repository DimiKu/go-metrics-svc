package handlers

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

func MetricReceiveAllMetricsHandler(service Service, log *zap.SugaredLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		type PageData struct {
			Items []string
		}
		tmpl := `
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
		<title>Metric collector</title>
	</head>
	<body>
		<h1>Metrics</h1>
		<h2>%s</h2>
	</body>
	</html>
    `
		var collectedMetricsString string
		metrics := service.GetAllMetrics()
		for _, metric := range metrics {
			metric = metric + "\n"
			collectedMetricsString += metric
		}

		rw.Header().Set("Content-Type", "text/html")
		page := fmt.Sprintf(tmpl, collectedMetricsString)
		rw.Write([]byte(page))
	}
}
