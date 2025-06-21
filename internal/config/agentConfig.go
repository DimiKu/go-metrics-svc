package config

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type AgentFileConfig struct {
	FlagRunAddr  string `json:"address"`
	PoolInterval string `json:"poll_interval"`
	SendInterval string `json:"report_interval"`
	UseCrypto    string `json:"crypto_key"`
}

type AgentFlagConfig struct {
	FlagRunAddr  string
	PoolInterval string
	SendInterval string
	UseHash      string
	WorkerCount  int
	UseCrypto    string
	ConfigPath   string
}

type AgentConfig struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
	UseHash        string `env:"KEY"`
	WorkerCount    int    `env:"RATE_LIMIT"`
	UseCrypto      string `env:"CRYPTO_KEY"`
	ConfigPath     string `env:"CONFIG"`
}

func ValidateAgentConfig(
	cfg AgentConfig,
	flagCfg AgentFlagConfig,
	log *zap.SugaredLogger,
) AgentConfig {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Error parse env: %s", err)
	}

	if cfg.Addr == "" {
		cfg.Addr = flagCfg.FlagRunAddr
	}

	if cfg.PollInterval == "" {
		cfg.PollInterval = flagCfg.PoolInterval
	}

	if cfg.ReportInterval == "" {
		cfg.ReportInterval = flagCfg.SendInterval
	}

	if cfg.UseCrypto == "" {
		cfg.UseCrypto = flagCfg.UseCrypto
	}

	if cfg.UseHash == "" {
		cfg.UseHash = flagCfg.UseHash
	}

	if cfg.WorkerCount == 0 {
		cfg.WorkerCount = flagCfg.WorkerCount
	}

	if cfg.ConfigPath != "" || flagCfg.ConfigPath != "" {
		if cfg.ConfigPath == "" && flagCfg.ConfigPath != "" {
			cfg.ConfigPath = flagCfg.ConfigPath
		}

		fileCfg := GetAgentConfigFromFile(cfg.ConfigPath, log)

		if cfg.Addr == "" {
			cfg.Addr = fileCfg.FlagRunAddr
		}

		if cfg.PollInterval == "" {
			cfg.PollInterval = fileCfg.PoolInterval
		}

		if cfg.ReportInterval == "" {
			cfg.ReportInterval = fileCfg.SendInterval
		}

		if cfg.UseCrypto == "" {
			cfg.UseCrypto = fileCfg.UseCrypto
		}
	}

	return cfg
}
