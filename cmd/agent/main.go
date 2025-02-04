package main

import (
	"github.com/caarlos0/env/v11"
	"go-metric-svc/internal/config"
	agentService "go-metric-svc/internal/service/agent"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func main() {
	var (
		cfg config.AgentConfig
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

	sugarLog.Infof("Start sending messages to %s", flagRunAddr)

	go agentService.PoolMetricsWorker(ch, time.Duration(poolDurationInterval), &counter)

	sendTicker := time.NewTicker(time.Duration(sendDurationInterval))
	defer sendTicker.Stop()

	go func() {
		for {
			<-sendTicker.C
			metrics := <-ch

			if err := agentService.SendMetrics(metrics, sugarLog, flagRunAddr); err != nil {
				sugarLog.Error(err)
			}
			sendTicker.Reset(time.Duration(sendDurationInterval))
		}
	}()

	for {
		sugarLog.Info("Agent tick")
		time.Sleep(1 * time.Second)
	}
}
