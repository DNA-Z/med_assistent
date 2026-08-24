package config

import "testing"

func TestGenericFunctionalOptions(t *testing.T) {
	config := NewConfig(
		WithDatabaseConnectionString("postgres://test"),
		WithTelegramToken("token"),
		WithProcessingWorkers(8),
		WithProcessingQueueSize(50),
	)

	if config.DBConnectionString != "postgres://test" {
		t.Fatalf("строка подключения=%q", config.DBConnectionString)
	}
	if config.Telegram.Token != "token" {
		t.Fatalf("токен=%q", config.Telegram.Token)
	}
	if config.Processing.Workers != 8 {
		t.Fatalf("количество обработчиков=%d", config.Processing.Workers)
	}
	if config.Processing.QueueSize != 50 {
		t.Fatalf("размер очереди=%d", config.Processing.QueueSize)
	}
}

func TestGenericApplySupportsAnyConfigurationType(t *testing.T) {
	type customConfig struct{ Enabled bool }
	value := customConfig{}
	Apply(&value, Option[customConfig](func(config *customConfig) { config.Enabled = true }))
	if !value.Enabled {
		t.Fatal("обобщённая функциональная опция не была применена")
	}
}

func TestOpenAIConfigurationFromEnvironment(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "openai")
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_MODEL", "gpt-test")
	t.Setenv("OPENAI_BASE_URL", "https://example.test/v1")

	config := NewConfig()
	config.applyEnvironment()

	if config.LLM.Provider != "openai" {
		t.Fatalf("провайдер=%q", config.LLM.Provider)
	}
	if config.LLM.OpenAI.APIKey != "test-key" {
		t.Fatalf("ключ API=%q", config.LLM.OpenAI.APIKey)
	}
	if config.LLM.OpenAI.Model != "gpt-test" {
		t.Fatalf("модель=%q", config.LLM.OpenAI.Model)
	}
	if config.LLM.OpenAI.BaseURL != "https://example.test/v1" {
		t.Fatalf("базовый URL=%q", config.LLM.OpenAI.BaseURL)
	}
}
