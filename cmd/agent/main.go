package main

import (
	"fmt"
	"go-metric-svc/internal/config"
	agentService "go-metric-svc/internal/service/agent"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

var (
	// Флаги, которые можно передать при компиляции
	// пример: go build -ldflags "-X main.buildVersion=1.0"
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	var (
		cfg            config.AgentConfig
		metricsMapLock sync.Mutex
		wg             sync.WaitGroup
	)

	config.GetBuildInfo(buildVersion, buildDate, buildCommit)

	type metricTransfer struct {
		Name  string
		Value float32
	}

	logger, _ := zap.NewProduction()
	sugarLog := logger.Sugar()

	parseFlags()
	cfg = config.ValidateAgentConfig(cfg, flags, sugarLog)

	metricChan := make(chan metricTransfer, cfg.WorkerCount)
	counter := 0
	metricsMap := make(map[string]float32)

	sugarLog.Infof("Pool intervar is %s", cfg.PollInterval)
	poolDurationInterval, err := strconv.Atoi(cfg.PollInterval)
	if err != nil {
		sugarLog.Error(err)
	}

	sugarLog.Infof("Send intervar is %s", cfg.ReportInterval)
	sendDurationInterval, err := strconv.Atoi(cfg.ReportInterval)
	if err != nil {
		sugarLog.Error(err)
	}

	sugarLog.Infof("Start sending messages to %s", cfg.Addr)

	poolTicker := time.NewTicker(time.Duration(poolDurationInterval) * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			<-poolTicker.C
			metricsMapLock.Lock()
			counter += 1
			metricsMap = agentService.PoolMetricsWorker(&counter)
			metricsMapLock.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			<-poolTicker.C
			metricsMapLock.Lock()
			counter += 1
			metricsMap = agentService.ExtraMetricsWorker(metricsMap)
			metricsMapLock.Unlock()
		}
	}()

	sendTicker := time.NewTicker(time.Duration(sendDurationInterval) * time.Second)
	defer sendTicker.Stop()

	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			<-sendTicker.C

			metricsMapLock.Lock()
			metrics := metricsMap
			metricsMapLock.Unlock()
			sugarLog.Infof("Start send metric in channel")

			for k, v := range metrics {
				metricChan <- metricTransfer{
					Name:  k,
					Value: v,
				}
			}
			counter = 0
		}
	}()

	wg.Add(1)
	go func() {
		for {
			for i := 0; i <= cfg.WorkerCount; i++ {
				metric, ok := <-metricChan
				if !ok {
					return
				}
				sugarLog.Infof("Start send metric")
				if err := agentService.SendJSONMetric(metric.Name, metric.Value, sugarLog, cfg.Addr, cfg.UseHash, cfg.UseCrypto); err != nil {
					fmt.Println("Error sending metric:", err)
				}
			}
		}
	}()

	for {
		sugarLog.Info("Agent tick")
		time.Sleep(1 * time.Second)
	}
}
