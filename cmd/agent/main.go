package main

import (
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var (
	poolInterval      = 2 * time.Second
	sendInterval      = 10 * time.Second
	host              = "http://localhost:8080"
	counterMetricName = "Counter"
)

func collectMetrics(counter *int) map[string]float32 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	rand.Seed(time.Now().UnixNano())

	metricsMap := map[string]float32{
		"Alloc":         float32(memStats.Alloc),
		"BuckHashSys":   float32(memStats.BuckHashSys),
		"Frees":         float32(memStats.Frees),
		"GCCPUFraction": float32(memStats.GCCPUFraction), // Приведение к float32
		"GCSys":         float32(memStats.GCSys),
		"HeapAlloc":     float32(memStats.HeapAlloc),
		"HeapIdle":      float32(memStats.HeapIdle),
		"HeapInuse":     float32(memStats.HeapInuse),
		"HeapObjects":   float32(memStats.HeapObjects),
		"HeapReleased":  float32(memStats.HeapReleased),
		"HeapSys":       float32(memStats.HeapSys),
		"LastGC":        float32(memStats.LastGC),
		"Lookups":       float32(memStats.Lookups),
		"MCacheInuse":   float32(memStats.MCacheInuse),
		"MCacheSys":     float32(memStats.MCacheSys),
		"MSpanInuse":    float32(memStats.MSpanInuse),
		"MSpanSys":      float32(memStats.MSpanSys),
		"Mallocs":       float32(memStats.Mallocs),
		"NextGC":        float32(memStats.NextGC),
		"NumForcedGC":   float32(memStats.NumForcedGC),
		"NumGC":         float32(memStats.NumGC),
		"OtherSys":      float32(memStats.OtherSys),
		"PauseTotalNs":  float32(memStats.PauseTotalNs),
		"StackInuse":    float32(memStats.StackInuse),
		"StackSys":      float32(memStats.StackSys),
		"Sys":           float32(memStats.Sys),
		"TotalAlloc":    float32(memStats.TotalAlloc),
		"RandomValue":   float32(rand.Float64()),
		"Counter":       float32(*counter),
	}
	return metricsMap
}

func poolMetricsWorker(ch chan map[string]float32, interval time.Duration, counter *int) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			*counter += 1
			metrics := collectMetrics(counter)
			ch <- metrics
		}
	}
}

func main() {
	ch := make(chan map[string]float32, 1)
	counter := 0
	logger, _ := zap.NewProduction()
	sugarLog := logger.Sugar()
	go poolMetricsWorker(ch, poolInterval, &counter)
	sendTicker := time.NewTicker(sendInterval)
	defer sendTicker.Stop()
	var url string
	go func() {
		for {
			select {
			case <-sendTicker.C:

				metrics := <-ch

				for k, v := range metrics {
					if k == counterMetricName {
						url = fmt.Sprintf("%s/update/%s/%s/%d", host, "gauge", k, int64(v))
					} else {
						url = fmt.Sprintf("%s/update/%s/%s/%f", host, "gauge", k, v)
					}

					sugarLog.Infof("Url is: %s", url)

					res, err := http.Post(url, "Content-Type: text/plain", nil)
					logger.Info(fmt.Sprintf("Send metric via url %s", url))
					defer res.Body.Close()
					if err != nil {
						sugarLog.Infof("Send metric via url %s", url)
					}
				}
			}

		}
	}()

	for {
		sugarLog.Info("Agent tick")
		time.Sleep(1 * time.Second)
	}
}
