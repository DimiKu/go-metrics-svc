package config

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
)

// GetBuildInfo ф-я выводит в STDOUT информацию о билде сборки
func GetBuildInfo(version, date, commit string) {
	values := []struct {
		name  string
		value string
	}{
		{"Build version", version},
		{"Build date", date},
		{"Build commit", commit},
	}

	for _, v := range values {
		if v.value == "" {
			v.value = "N/A"
		}
		fmt.Printf("%s: %s\n", v.name, v.value)
	}
}

func GetServerConfigFromFile(path string, log *zap.SugaredLogger) ServerFileConfig {
	var cfg ServerFileConfig

	data, err := os.ReadFile(path)
	if err != nil {
		log.Infof("Error reading config file: %s", err)
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Infof("Error parsing config file: %s", err)
	}

	return cfg
}

func GetAgentConfigFromFile(path string, log *zap.SugaredLogger) AgentFileConfig {
	var cfg AgentFileConfig

	data, err := os.ReadFile(path)
	if err != nil {
		log.Infof("Error reading config file: %s", err)
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Infof("Error parsing config file: %s", err)
	}

	return cfg
}
