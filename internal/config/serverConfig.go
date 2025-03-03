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
}

func ValidateServerConfig(
	cfg ServerConfig,
	flagRunAddr string,
	storeInterval string,
	fileStoragePath string,
) (string, string, string) {
	var addr, saveInterval, filePathToStoreMetrics string

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

	return addr, saveInterval, filePathToStoreMetrics
}
