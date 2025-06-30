package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go-metric-svc/internal/entities/agent"
	"go-metric-svc/internal/models"
	"go-metric-svc/internal/utils"
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
		"GCCPUFraction": float32(memStats.GCCPUFraction),
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

func PoolMetricsWorker(counter *int) map[string]float32 {
	metrics := collectMetrics(counter)
	return metrics
}

func ExtraMetricsWorker(metricsMap map[string]float32) map[string]float32 {
	var (
		totalMemory     float32
		freeMemory      float32
		cpuUtilization1 float32
	)

	vmStat, err := mem.VirtualMemory()
	if err == nil {
		freeMemory = float32(vmStat.Free)
		totalMemory = float32(vmStat.Total)
	}

	cpuStats, err := cpu.Percent(0, false)
	if err == nil && len(cpuStats) > 0 {
		cpuUtilization1 = float32(cpuStats[0])
	}

	metricsMap["TotalMemory"] = totalMemory
	metricsMap["FreeMemory"] = freeMemory
	metricsMap["CPUutilization1"] = cpuUtilization1

	return metricsMap
}

func SendMetrics(metricsMap map[string]float32, log *zap.SugaredLogger, host string) error {
	var url string
	log.Info("start send metrics")
	hostWithSchema := "https://" + host
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

func SendJSONMetric(metricType string, metricValue float32, log *zap.SugaredLogger, host string, useHash string, useCrypto string) error {
	shema := "http://"
	var metric models.Metrics
	var b bytes.Buffer

	if metricType == agent.CounterMetricName {
		metric.ID = metricType
		metric.MType = agent.CounterMetricType
		value := int64(metricValue)
		metric.Delta = &value
	} else {
		metric.ID = metricType
		metric.MType = agent.GaugeMetricName
		value := float64(metricValue)
		metric.Value = &value
	}

	resMetrics, err := json.Marshal(metric)
	if err != nil {
		log.Infof("Failed to marshal body")
		return err
	}

	// TODO не понял почему пишу два раза. Надо обсудить
	w := gzip.NewWriter(&b)

	if useCrypto != "" {
		data, err := utils.EncryptWithCert(useCrypto, resMetrics)
		if err != nil {
			log.Infof("Failed to encrypt data")
		}
		w.Write(data)
	}

	_, err = w.Write(resMetrics)
	if err != nil {
		fmt.Println("Error writing gzip data:", err)
		return err
	}

	url := shema + host + "/update/"

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

	if useHash != "" {
		h := hmac.New(sha256.New, []byte(useHash))
		h.Write(resMetrics)
		hashBytes := h.Sum(nil)
		hashString := hex.EncodeToString(hashBytes)
		req.Header.Set("HashSHA256", hashString)
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(b.Len())

	log.Infof("Do req: %s", req.Body)
	err = doReqWithRetry(*req, *log)
	if err != nil {
		log.Errorf("Error in send metrics: %s", err)
	}

	return nil
}

func doReqWithRetry(req http.Request, log zap.SugaredLogger) error {
	for i := 1; i < 6; i += 2 {
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
		time.Sleep(time.Duration(i))
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
