package config

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type ServerFileConfig struct {
	Addr          string `json:"address"`
	Restore       bool   `json:"base_url"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DBDsn         string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

type ServerFlagConfig struct {
	FlagRunAddr     string
	StoreInterval   string
	FileStoragePath string
	NeedRestore     bool
	UseHash         string
	ConnString      string
	UseCrypto       string
	ConfigPath      string
	TrustedSubnet   string
	GRPCAddr        string
}

type ServerConfig struct {
	Addr            string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	NeedRestore     bool   `env:"RESTORE"`
	StorageInterval string `env:"STORE_INTERVAL"`
	ConnString      string `env:"DATABASE_DSN"`
	UseHash         string `env:"KEY"`
	UseCrypto       string `env:"CRYPTO_KEY"`
	ConfigPath      string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
	GRPCAddr        string `env:"GRPC_ADDR"`
}

func ValidateServerConfig(cfg ServerConfig, flagCfg ServerFlagConfig, log *zap.SugaredLogger) (ServerConfig, error) {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Error parse env: %s", err)
	}

	if cfg.Addr == "" {
		cfg.Addr = flagCfg.FlagRunAddr
	}

	if cfg.StorageInterval == "" {
		cfg.StorageInterval = flagCfg.StoreInterval
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = flagCfg.FileStoragePath
	}

	if cfg.ConnString == "" {
		cfg.ConnString = flagCfg.ConnString
	}

	if cfg.UseHash == "" {
		cfg.UseHash = flagCfg.UseHash
	}

	if cfg.UseCrypto == "" {
		cfg.UseCrypto = flagCfg.UseCrypto
	}

	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = flagCfg.TrustedSubnet
	}

	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = flagCfg.GRPCAddr
	}

	if cfg.ConfigPath != "" || flagCfg.ConfigPath != "" {
		if cfg.ConfigPath == "" && flagCfg.ConfigPath != "" {
			cfg.ConfigPath = flagCfg.ConfigPath
		}

		fileCfg := GetServerConfigFromFile(cfg.ConfigPath, log)

		if cfg.Addr == "" {
			cfg.Addr = fileCfg.Addr
		}

		if cfg.StorageInterval == "" {
			cfg.StorageInterval = fileCfg.StoreInterval
		}

		if cfg.FileStoragePath == "" {
			cfg.FileStoragePath = fileCfg.StoreFile
		}

		if cfg.ConnString == "" {
			cfg.ConnString = fileCfg.DBDsn
		}

		if cfg.UseCrypto == "" {
			cfg.UseCrypto = flagCfg.UseCrypto
		}

		if cfg.TrustedSubnet == "" {
			cfg.TrustedSubnet = flagCfg.TrustedSubnet
		}
	}

	return cfg, nil
}
