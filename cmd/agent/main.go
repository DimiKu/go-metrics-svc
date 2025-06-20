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

	metricChan := make(chan metricTransfer, workerCount)
	counter := 0
	metricsMap := make(map[string]float32)

	parseFlags()
	logger, _ := zap.NewProduction()
	sugarLog := logger.Sugar()

	poolInterval, sendInterval, flagRunAddr, useHash, workerCount, useCrypto = config.ValidateAgentConfig(cfg, flagRunAddr, poolInterval, sendInterval, useHash, workerCount, useCrypto)

	sugarLog.Infof("Pool intervar is %s", poolInterval)
	poolDurationInterval, err := strconv.Atoi(poolInterval)
	if err != nil {
		sugarLog.Error(err)
	}

	sugarLog.Infof("Send intervar is %s", sendInterval)
	sendDurationInterval, err := strconv.Atoi(sendInterval)
	if err != nil {
		sugarLog.Error(err)
	}

	sugarLog.Infof("Start sending messages to %s", flagRunAddr)

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
			for i := 0; i <= workerCount; i++ {
				metric, ok := <-metricChan
				if !ok {
					return
				}
				sugarLog.Infof("Start send metric")
				if err := agentService.SendJSONMetric(metric.Name, metric.Value, sugarLog, flagRunAddr, useHash, useCrypto); err != nil {
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
