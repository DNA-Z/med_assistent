package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(
	addr string,
	password string,
	db int,
) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
