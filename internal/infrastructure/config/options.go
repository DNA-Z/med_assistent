package config

// Option is a reusable generic functional option for configuration values.
type Option[T any] func(*T)

// Apply applies options in declaration order; later options override earlier ones.
func Apply[T any](target *T, options ...Option[T]) {
	for _, option := range options {
		if option != nil {
			option(target)
		}
	}
}

func WithDatabaseConnectionString(value string) Option[Config] {
	return func(config *Config) { config.DBConnectionString = value }
}

func WithTelegramToken(value string) Option[Config] {
	return func(config *Config) { config.Telegram.Token = value }
}

func WithProcessingWorkers(value int) Option[Config] {
	return func(config *Config) {
		if value > 0 {
			config.Processing.Workers = value
		}
	}
}
