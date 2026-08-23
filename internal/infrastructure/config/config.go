// Package config отвечает за загрузку и проверку конфигурации сервиса.
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
	ObjectStorage struct {
		Endpoint  string `yaml:"endpoint"`
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
		Bucket    string `yaml:"bucket"`
		UseSSL    bool   `yaml:"use_ssl"`
	} `yaml:"object_storage"`
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
		Workers   int `yaml:"workers"`
		QueueSize int `yaml:"queue_size"`
		Timeout   int `yaml:"timeout"`
	} `yaml:"processing"`
}

// NewConfig возвращает конфигурацию с безопасными значениями по умолчанию
// и применяет переданные функциональные опции.
func NewConfig(options ...Option[Config]) *Config {
	c := &Config{}
	c.Redis.Address = "localhost:6379"
	c.ObjectStorage.Endpoint = "localhost:9000"
	c.ObjectStorage.Bucket = "medical-audio"
	c.Telegram.Timeout = 10
	c.Speech.Provider, c.Speech.Timeout = "mock", 60
	c.LLM.Provider, c.LLM.Model, c.LLM.Timeout = "mock", "GigaChat-2", 60
	c.Processing.Workers, c.Processing.QueueSize, c.Processing.Timeout = 4, 100, 900
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
	set("S3_ENDPOINT", &o.ObjectStorage.Endpoint)
	set("S3_ACCESS_KEY", &o.ObjectStorage.AccessKey)
	set("S3_SECRET_KEY", &o.ObjectStorage.SecretKey)
	set("S3_BUCKET", &o.ObjectStorage.Bucket)
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
	if value := os.Getenv("S3_USE_SSL"); value != "" {
		if enabled, err := strconv.ParseBool(value); err == nil {
			o.ObjectStorage.UseSSL = enabled
		}
	}
	if value := os.Getenv("PROCESSING_WORKERS"); value != "" {
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			o.Processing.Workers = n
		}
	}
	if value := os.Getenv("PROCESSING_QUEUE_SIZE"); value != "" {
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			o.Processing.QueueSize = n
		}
	}
}
