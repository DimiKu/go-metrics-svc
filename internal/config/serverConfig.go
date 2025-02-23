package config

type ServerConfig struct {
	Addr            string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	NeedRestore     bool   `env:"RESTORE"`
	StorageInterval string `env:"STORE_INTERVAL"`
}
