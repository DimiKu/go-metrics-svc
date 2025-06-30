package main

import (
	"context"
	"fmt"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

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
			select {
			case <-ctx.Done():
				fmt.Println("Shutting down pool goroutines")
				return
			default:
				<-poolTicker.C
				metricsMapLock.Lock()
				counter += 1
				metricsMap = agentService.PoolMetricsWorker(&counter)
				metricsMapLock.Unlock()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Shutting down extra worker goroutines")
				return
			default:
				<-poolTicker.C
				metricsMapLock.Lock()
				counter += 1
				metricsMap = agentService.ExtraMetricsWorker(metricsMap)
				metricsMapLock.Unlock()
			}
		}
	}()

	sendTicker := time.NewTicker(time.Duration(sendDurationInterval) * time.Second)
	defer sendTicker.Stop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Shutting down collect metric in channel goroutines")
				return
			default:
				<-sendTicker.C
				metricsMapLock.Lock()
				metrics := metricsMap
				metricsMapLock.Unlock()
				sugarLog.Infof("Start collect metric in channel")

				for k, v := range metrics {
					metricChan <- metricTransfer{
						Name:  k,
						Value: v,
					}
				}
				counter = 0
				sendTicker.Reset(time.Duration(sendDurationInterval) * time.Second)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Shutting down send goroutines")
				return
			default:
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
		}
	}()

	<-signalChan
	fmt.Println("Agent start graceful Shutdown")
	cancel()

	// TODO тут должен быть wg.Wait() но не получается. что-то блокируется. Надо обсудить
	fmt.Println("Agent Shutdown gracefully")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down send goroutines")
			return
		default:
			sugarLog.Info("Agent tick")
			time.Sleep(1 * time.Second)
		}
	}
}
