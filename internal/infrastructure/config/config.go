// Package config предоставляет конфигурацию для сервиса.
package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config представляет структуру конфигурационного файла.
type Config struct {
	DBConnectionString string `yaml:"db_connection_string"`
	Telegram           struct {
		Token   string `yaml:"token"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"telegram"`
	Redis struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	Speech struct {
		Provider string `yaml:"provider"`
		APIKey   string `yaml:"api_key"`
		BaseURL  string `yaml:"base_url"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"speech"`
	LLM struct {
		Provider string `yaml:"provider"`
		Token    string `yaml:"token"`
		Model    string `yaml:"model"`
		BaseURL  string `yaml:"base_url"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"llm"`
	Processing struct {
		Workers int `yaml:"workers"`
		Timeout int `yaml:"timeout"`
	} `yaml:"processing"`
}

// NewConfig возвращает конфигурацию с безопасными значениями по умолчанию
// и применяет переданные функциональные опции.
func NewConfig(options ...Option[Config]) *Config {
	c := &Config{}
	c.Redis.Address = "localhost:6379"
	c.Telegram.Timeout = 10
	c.Speech.Provider, c.Speech.Timeout = "mock", 60
	c.LLM.Provider, c.LLM.Model, c.LLM.Timeout = "mock", "GigaChat-2", 60
	c.Processing.Workers, c.Processing.Timeout = 4, 900
	Apply(c, options...)
	return c
}

// ConfigInit загружает YAML-конфигурацию и переопределяет её переменными окружения.
func (o *Config) ConfigInit() error {
	cfg, err := o.readConfigFile()
	if err != nil {
		return err
	}
	*o = *cfg
	o.applyEnvironment()
	return nil
}

func (o *Config) readConfigFile() (*Config, error) {
	configFile := os.Getenv("MED_ASSISTANT_CONFIG")
	if configFile == "" {
		configFile = "internal/infrastructure/config/config.yaml"
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read config file %s: %w",
			configFile,
			err,
		)
	}

	cfg := NewConfig()

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf(
			"failed to parse config file %s: %w",
			configFile,
			err,
		)
	}

	return cfg, nil
}

func (o *Config) applyEnvironment() {
	set := func(key string, dst *string) {
		if value := os.Getenv(key); value != "" {
			*dst = value
		}
	}
	set("DATABASE_URL", &o.DBConnectionString)
	set("REDIS_ADDRESS", &o.Redis.Address)
	set("REDIS_PASSWORD", &o.Redis.Password)
	set("TELEGRAM_TOKEN", &o.Telegram.Token)
	set("SPEECH_PROVIDER", &o.Speech.Provider)
	set("SPEECH_API_KEY", &o.Speech.APIKey)
	set("LLM_PROVIDER", &o.LLM.Provider)
	set("GIGACHAT_TOKEN", &o.LLM.Token)
	if value := os.Getenv("REDIS_DB"); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			o.Redis.DB = n
		}
	}
}
