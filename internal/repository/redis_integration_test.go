//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/suprt/currency_converter/internal/service"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisContainer(t *testing.T) (*RedisRepository, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	host, err := redisC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get the redis host: %v", err)
	}
	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get the redis port: %v", err)
	}

	repo, err := NewRedisStorage(host+":"+port.Port(), "", 0)
	if err != nil {
		t.Fatalf("failed to init redis storage: %v", err)
	}

	cleanup := func() {
		_ = repo.Close()
		_ = redisC.Terminate(ctx)
	}
	return repo, cleanup
}

func TestRedisIntegration_SetAndGet(t *testing.T) {
	repo, cleanup := setupRedisContainer(t)
	defer cleanup()

	ctx := context.Background()

	err := repo.Set(ctx, "USD:EUR", 0.85, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := repo.Get(ctx, "USD:EUR")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != 0.85 {
		t.Fatalf("Expected 0.85, got %f", val)
	}
}

func TestRedisIntegration_TTL(t *testing.T) {
	repo, cleanup := setupRedisContainer(t)
	defer cleanup()
	ctx := context.Background()
	err := repo.Set(ctx, "USD:EUR", 0.85, 1*time.Second)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	time.Sleep(2 * time.Second)
	val, err := repo.Get(ctx, "USD:EUR")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("Expected error, got nil")
	}
	if val != 0 {
		t.Fatalf("Expected 0, got %f", val)
	}
}

func TestRedisIntegration_Delete(t *testing.T) {
	repo, cleanup := setupRedisContainer(t)
	defer cleanup()
	ctx := context.Background()
	err := repo.Set(ctx, "USD:EUR", 0.85, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	err = repo.Delete(ctx, "USD:EUR")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	val, err := repo.Get(ctx, "USD:EUR")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("Expected error, got nil")
	}
	if val != 0 {
		t.Fatalf("Expected 0, got %f", val)
	}
}

func TestRedisIntegration_Clear(t *testing.T) {
	repo, cleanup := setupRedisContainer(t)
	defer cleanup()
	ctx := context.Background()
	keys := []struct {
		key string
		val float64
	}{
		{key: "USD:EUR", val: 0.85},
		{key: "USD:RUB", val: 80},
		{key: "BTC:USD", val: 100000},
		{key: "ETH:USD", val: 2000},
		{key: "EUR:USD", val: 1.1},
	}
	for _, key := range keys {
		err := repo.Set(ctx, key.key, key.val, 0)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}
	val, err := repo.Get(ctx, "USD:EUR")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != keys[0].val {
		t.Fatalf("Expected %f, got %f", keys[0].val, val)
	}
	err = repo.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	val, err = repo.Get(ctx, "USD:EUR")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("Expected error, got nil")
	}
	if val != 0 {
		t.Fatalf("Expected 0, got %f", val)
	}
}
