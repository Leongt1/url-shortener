package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("url not found")

type Repository struct {
	client *redis.Client
}

func NewRepository(addr string) (*Repository, error) {
	client := redis.NewClient(
		&redis.Options{
			Addr: addr,
		},
	)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to create new redis instance: %w", err)
	}

	return &Repository{
		client: client,
	}, nil
}

func (r *Repository) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error pinging redis: %w", err)
	}

	return nil
}

func (r *Repository) SetUrl(ctx context.Context, code string, value string, expiration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, "url:"+code, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("error setting url: %w", err)
	}

	return ok, nil
}

func (r *Repository) GetUrl(ctx context.Context, code string) (string, error) {
	url, err := r.client.Get(ctx, "url:"+code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}

	return url, nil
}
