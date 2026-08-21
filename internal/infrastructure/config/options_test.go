package config

import "testing"

func TestGenericFunctionalOptions(t *testing.T) {
	config := NewConfig(
		WithDatabaseConnectionString("postgres://test"),
		WithTelegramToken("token"),
		WithProcessingWorkers(8),
	)

	if config.DBConnectionString != "postgres://test" {
		t.Fatalf("dsn=%q", config.DBConnectionString)
	}
	if config.Telegram.Token != "token" {
		t.Fatalf("token=%q", config.Telegram.Token)
	}
	if config.Processing.Workers != 8 {
		t.Fatalf("workers=%d", config.Processing.Workers)
	}
}

func TestGenericApplySupportsAnyConfigurationType(t *testing.T) {
	type customConfig struct{ Enabled bool }
	value := customConfig{}
	Apply(&value, Option[customConfig](func(config *customConfig) { config.Enabled = true }))
	if !value.Enabled {
		t.Fatal("generic option was not applied")
	}
}
