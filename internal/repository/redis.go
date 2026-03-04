package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/suprt/currency_converter/internal/service"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisStorage(addr, password string, db int) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisRepository{client: client}, nil

}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}

func (r *RedisRepository) Get(ctx context.Context, key string) (float64, error) {
	val, err := r.client.Get(ctx, key).Float64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, service.ErrNotFound
		}
		return 0, fmt.Errorf("redis get failed: %w", err)
	}

	return val, nil

}

func (r *RedisRepository) Set(ctx context.Context, key string, val float64, ttl time.Duration) error {
	err := r.client.Set(ctx, key, val, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	val, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return val > 0, nil
}

func (r *RedisRepository) TTL(ctx context.Context, key string) (time.Duration, error) {

	val, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}
	if val == -2 {
		return 0, service.ErrNotFound
	}
	return val, nil
}

func (r *RedisRepository) Len(ctx context.Context) (int64, error) {
	var count int64
	iter := r.client.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("redis scan failed: %w", err)
	}
	return count, nil
}

func (r *RedisRepository) Clear(ctx context.Context) error {
	var keys []string
	iter := r.client.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis scan failed: %w", err)
	}
	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}
