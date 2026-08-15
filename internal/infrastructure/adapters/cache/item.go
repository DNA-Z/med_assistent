package cache

import "time"

type Item struct {
	Value     []byte
	ExpiresAt time.Time
}
