package cache

import (
	"context"
	"sync"
	"time"
)

type MemoryCache struct {
	mu      sync.RWMutex
	items   map[string]Item
	maxSize int
	stop    chan struct{}
	done    chan struct{}
}

func (c *MemoryCache) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.RLock()

	item, ok := c.items[key]

	c.mu.RUnlock()

	if !ok {
		return nil, ports.ErrCacheMiss
	}

	if !item.ExpiresAt.IsZero() &&
		time.Now().After(item.ExpiresAt) {

		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()

		return nil, ports.ErrCacheMiss
	}

	return item.Value, nil
}

func (c *MemoryCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	ttl time.Duration,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	var expiresAt time.Time

	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	return nil
}

func (c *MemoryCache) Delete(
	ctx context.Context,
	key string,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)

	return nil
}

func (c *MemoryCache) Close() {
	close(c.stop)
	<-c.done
}
