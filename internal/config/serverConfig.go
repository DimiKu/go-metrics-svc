package config

import (
	"github.com/caarlos0/env/v11"
	"log"
)

type ServerConfig struct {
	Addr            string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	NeedRestore     bool   `env:"RESTORE"`
	StorageInterval string `env:"STORE_INTERVAL"`
	ConnString      string `env:"DATABASE_DSN"` // 'postgres://myuser:metricpass@postgres:5432/metric_db?sslmode=disable'
}

func ValidateServerConfig(
	cfg ServerConfig,
	flagRunAddr string,
	storeInterval string,
	fileStoragePath string,
	connectionString string,
) (string, string, string, string) {
	var addr, saveInterval, filePathToStoreMetrics, connString string

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Error parse env: %s", err)
	}

	if cfg.Addr != "" {
		addr = cfg.Addr
	} else {
		addr = flagRunAddr
	}

	if cfg.StorageInterval != "" {
		saveInterval = cfg.StorageInterval
	} else {
		saveInterval = storeInterval
	}

	if cfg.FileStoragePath != "" {
		filePathToStoreMetrics = cfg.FileStoragePath
	} else {
		filePathToStoreMetrics = fileStoragePath
	}

	if cfg.ConnString != "" {
		connString = cfg.ConnString
	} else {
		connString = connectionString
	}

	return addr, saveInterval, filePathToStoreMetrics, connString
}
