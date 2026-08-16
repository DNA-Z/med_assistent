// Package config предоставляет конфигурацию для сервиса.
package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config представляет структуру конфигурационного файла.
type Config struct {
	ServerAddress      string `yaml:"server_address"`
	BaseURL            string `yaml:"base_url"`
	DBConnectionString string `yaml:"db_connection_string"`
	SecretKey          string `yaml:"secret_key"`
	TgBotToken         string `yaml:"tg_bot_token"`
	WhisperAPIKey      string `yaml:"whisper_api_key"`
	GigaChatToken      string `yaml:"gigachat_token"`
	GigaChatModel      string `yaml:"gigachat_model"`
}

func NewConfig() *Config {
	return &Config{
		ServerAddress:      "",
		BaseURL:            "",
		DBConnectionString: "",
		SecretKey:          "",
		TgBotToken:         "",
		WhisperAPIKey:      "",
		GigaChatToken:      "",
		GigaChatModel:      "GigaChat-2",
	}
}

func (o *Config) ConfigInit() {
	cfg, err := o.readConfigFile()
	if err != nil {
		log.Printf("Warning: failed to read config file: %v", err)
		return
	}

	*o = *cfg
}

func (o *Config) readConfigFile() (*Config, error) {
	configFile := "config.yaml"

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read config file %s: %w",
			configFile,
			err,
		)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf(
			"failed to parse config file %s: %w",
			configFile,
			err,
		)
	}

	return &cfg, nil
}
