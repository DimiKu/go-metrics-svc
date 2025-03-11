package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go-metric-svc/internal/entities/agent"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

func collectMetrics(counter *int) map[string]float32 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	pollCount := *counter

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
		"PollCount":     float32(pollCount),
	}
	return metricsMap
}

func PoolMetricsWorker(ch chan map[string]float32, counter *int) map[string]float32 {
	metrics := collectMetrics(counter)
	return metrics
	//ch <- metrics
}

func SendMetrics(metricsMap map[string]float32, log *zap.SugaredLogger, host string) error {
	var url string
	log.Info("start send metrics")
	hostWithSchema := "http://" + host
	for k, v := range metricsMap {
		if k == agent.CounterMetricName {
			url = fmt.Sprintf("%s/update/%s/%s/%d", hostWithSchema, "counter", k, int64(v))
		} else {
			url = fmt.Sprintf("%s/update/%s/%s/%f", hostWithSchema, "gauge", k, v)
		}
		res, err := http.Post(url, "Content-Type: text/plain", nil)
		if err != nil {
			log.Infof("Send metric via url %s", url)
			return err
		}
		defer res.Body.Close()
	}

	return nil
}

func SendJSONMetrics(metricsMap map[string]float32, log *zap.SugaredLogger, host string) error {
	url := "http://" + host + "/update/"
	for k, v := range metricsMap {
		var metric models.Metrics
		if k == agent.CounterMetricName {
			metric.ID = k
			metric.MType = agent.CounterMetricType
			value := int64(v)
			metric.Delta = &value
		} else {
			metric.ID = k
			metric.MType = agent.GaugeMetricName
			value := float64(v)
			metric.Value = &value
		}

		resMetrics, err := json.Marshal(metric)
		if err != nil {
			log.Infof("Failed to marshal body %s", url)
			return err
		}
		var b bytes.Buffer

		w := gzip.NewWriter(&b)
		_, err = w.Write(resMetrics)
		if err != nil {
			fmt.Println("Error writing gzip data:", err)
			return err
		}

		err = w.Close()
		if err != nil {
			fmt.Println("Error closing gzip writer:", err)
			return err
		}

		req, err := http.NewRequest("POST", url, &b)
		if err != nil {
			log.Infof("Send metric via url %s", url)
			return err
		}

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(b.Len())

		err = doReqWithRetry(3, *req, *log)
		if err != nil {
			//return err
			log.Errorf("Error in send metrics: %s", err)
		}
		//client := &http.Client{}
		//resp, err := client.Do(req)
		//if err != nil {
		//	fmt.Println("Error sending request:", err)
		//	return nil
		//}
		//defer resp.Body.Close()

		// для дебага
		//body, err := io.ReadAll(resp.Body)
		//if err != nil {
		//	fmt.Println("Error reading response:", err)
		//	return nil
		//}
		//fmt.Println("Response Status:", resp.Status)
		//fmt.Println("Response Body:", string(body))
	}

	return nil
}

func doReqWithRetry(retry int, req http.Request, log zap.SugaredLogger) error {
	for i := 1; i < retry; i++ {
		client := &http.Client{}
		resp, err := client.Do(&req)
		if resp != nil && isRetryableStatusCode(resp.StatusCode) {
			continue
		}
		if err != nil {
			log.Infof("Error sending request:", err)
			return nil
		}
		defer resp.Body.Close()
		return nil
	}
	return nil
}

func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusInternalServerError, // 500
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout,     // 504
		http.StatusRequestTimeout,     // 408
		http.StatusTooManyRequests:    // 429
		return true
	}
	return false
}
