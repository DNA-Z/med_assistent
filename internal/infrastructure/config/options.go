package config

// Option описывает обобщённую функциональную опцию для настройки значения T.
type Option[T any] func(*T)

// Apply применяет опции по порядку объявления: последующие опции могут
// переопределить значения, установленные предыдущими.
func Apply[T any](target *T, options ...Option[T]) {
	for _, option := range options {
		if option != nil {
			option(target)
		}
	}
}

// WithDatabaseConnectionString задаёт строку подключения к PostgreSQL.
func WithDatabaseConnectionString(value string) Option[Config] {
	return func(config *Config) { config.DBConnectionString = value }
}

// WithTelegramToken задаёт токен Telegram-бота.
func WithTelegramToken(value string) Option[Config] {
	return func(config *Config) { config.Telegram.Token = value }
}

// WithProcessingWorkers задаёт положительный предел параллельной обработки.
func WithProcessingWorkers(value int) Option[Config] {
	return func(config *Config) {
		if value > 0 {
			config.Processing.Workers = value
		}
	}
}

// WithProcessingQueueSize задаёт положительную вместимость очереди обработки.
func WithProcessingQueueSize(value int) Option[Config] {
	return func(config *Config) {
		if value > 0 {
			config.Processing.QueueSize = value
		}
	}
}
