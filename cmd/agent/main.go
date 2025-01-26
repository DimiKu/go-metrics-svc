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
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

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
		"RandomValue":   float32(r.Float64()),
		"Counter":       float32(*counter),
	}
	return metricsMap
}

func poolMetricsWorker(ch chan map[string]float32, interval time.Duration, counter *int) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	<-ticker.C
	*counter += 1
	metrics := collectMetrics(counter)
	ch <- metrics

	return nil
}

func sendMetrics(metricsMap map[string]float32, log *zap.SugaredLogger) error {
	var url string

	for k, v := range metricsMap {
		if k == counterMetricName {
			url = fmt.Sprintf("%s/update/%s/%s/%d", host, "counter", k, int64(v))
		} else {
			url = fmt.Sprintf("%s/update/%s/%s/%f", host, "gauge", k, v)
		}

		log.Infof("Url is: %s", url)

		log.Info(fmt.Sprintf("Send metric via url %s", url))

		res, err := http.Post(url, "Content-Type: text/plain", nil)
		defer res.Body.Close()
		if err != nil {
			log.Infof("Send metric via url %s", url)
			return err
		}
	}

	return nil
}

func main() {
	ch := make(chan map[string]float32, 1)
	counter := 0

	logger, _ := zap.NewProduction()
	sugarLog := logger.Sugar()

	go poolMetricsWorker(ch, poolInterval, &counter)

	sendTicker := time.NewTicker(sendInterval)
	defer sendTicker.Stop()

	go func() {
		<-sendTicker.C
		metrics := <-ch

		if err := sendMetrics(metrics, sugarLog); err != nil {
			sugarLog.Error(err)
		}
	}()

	for {
		sugarLog.Info("Agent tick")
		time.Sleep(1 * time.Second)
	}
}
