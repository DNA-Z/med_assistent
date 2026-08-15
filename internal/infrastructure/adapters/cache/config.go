package cache

import "time"

type Config struct {
	MaxSize         int
	CleanupInterval time.Duration
	DefaultTTL      time.Duration
}
