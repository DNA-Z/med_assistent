// Package config предоставляет конфигурацию для сервиса
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Config представляет структуру конфигурационного файла.
type Config struct {
	ServerAddress      string `json:"server_address"`
	BaseURL            string `json:"base_url"`
	DBConnectionString string `json:"db_connection_string"`
	SecretKey          string `json:"secret_key"`
	TgBotToken         string `json:"tg_bot_token"`
}

func NewConfig() *Config {
	return &Config{
		ServerAddress:      "",
		BaseURL:            "",
		DBConnectionString: "",
		SecretKey:          "",
	}
}

func (o *Config) ConfigInit() {
	jsonConfig, err := o.readConfigFile()
	if err != nil {
		log.Printf("Warning: failed to read config file: %v", err)
	}
}

func (o *Config) readConfigFile() (*Config, error) {
	configFile := "config.json"

	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", configFile, err)
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	data := buf.Bytes()

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	return &cfg, nil
}
