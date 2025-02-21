package main

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"go-metric-svc/internal/config"
	agentService "go-metric-svc/internal/service/agent"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	var (
		cfg            config.AgentConfig
		metricsMapLock sync.Mutex
		metricsMap     map[string]float32
	)
	ch := make(chan map[string]float32, 1)
	counter := 0

	parseFlags()
	logger, _ := zap.NewProduction()
	sugarLog := logger.Sugar()

	err := env.Parse(&cfg)
	if err != nil {
		sugarLog.Errorf("Error parse env: %s", err)
	}

	if cfg.Addr != "" {
		flagRunAddr = cfg.Addr
	}

	if cfg.PollInterval != "" {
		poolInterval = cfg.PollInterval
	}

	if cfg.ReportInterval != "" {
		sendInterval = cfg.ReportInterval
	}

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sugarLog.Infof("Start sending messages to %s", flagRunAddr)

	poolTicker := time.NewTicker(time.Duration(poolDurationInterval) * time.Second)

	go func() {
		for {
			<-poolTicker.C
			metricsMapLock.Lock()
			counter += 1
			metricsMap = agentService.PoolMetricsWorker(ch, &counter)
			metricsMapLock.Unlock()
		}
	}()

	sendTicker := time.NewTicker(time.Duration(sendDurationInterval) * time.Second)
	defer sendTicker.Stop()

	go func() {
		for {
			<-sendTicker.C

			metricsMapLock.Lock()
			metrics := metricsMap
			metricsMapLock.Unlock()

			if err := agentService.SendJSONMetrics(metrics, sugarLog, flagRunAddr); err != nil {
				sugarLog.Error(err)
			}
		}
	}()

	select {
	case sig := <-sigChan:
		fmt.Println("Получен сигнал:", sig)
		os.Exit(0)
	}
}
