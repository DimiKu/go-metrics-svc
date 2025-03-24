package config

import (
	"github.com/caarlos0/env/v11"
	"log"
)

type AgentConfig struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
	UseHash        string `env:"KEY"`
}

func ValidateAgentConfig(
	cfg AgentConfig,
	flagRunAddr string,
	poolInterval string,
	sendInterval string,
	useHash string,
) (string, string, string, string) {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Error parse env: %s", err)
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

	if cfg.UseHash != "" {
		useHash = cfg.UseHash
	}

	return poolInterval, sendInterval, flagRunAddr, useHash
}
