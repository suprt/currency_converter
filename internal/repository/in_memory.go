package repository

import (
	"context"
	"sync"
	"time"

	"github.com/suprt/currency_converter/internal/service"
)

type InMemoryRepository struct {
	mu        sync.RWMutex
	items     map[string]item
	stopClean chan struct{}
}
type item struct {
	value     float64
	expiresAt int64
}

func NewInMemoryRepository(cleanupInterval time.Duration) *InMemoryRepository {
	ms := &InMemoryRepository{
		items:     make(map[string]item),
		stopClean: make(chan struct{}),
	}
	if cleanupInterval > 0 {
		ms.startCleanup(cleanupInterval)
	}
	return ms
}

func (r *InMemoryRepository) Set(ctx context.Context, key string, value float64, ttl time.Duration) error {

	var expiresAt int64

	r.mu.Lock()
	defer r.mu.Unlock()

	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).Unix()
	}
	r.items[key] = item{
		value:     value,
		expiresAt: expiresAt,
	}

	return nil
}

func (r *InMemoryRepository) Get(ctx context.Context, key string) (float64, error) {
	r.mu.RLock()
	it, ok := r.items[key]
	r.mu.RUnlock()

	if !ok {
		return 0, service.ErrNotFound
	}

	if it.expiresAt > 0 && it.expiresAt < time.Now().Unix() {

		r.mu.Lock()
		delete(r.items, key)
		r.mu.Unlock()
		return 0, service.ErrNotFound
	}

	return it.value, nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, key string) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, key)
	return nil
}

func (r *InMemoryRepository) Exists(ctx context.Context, key string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.items[key]
	return ok, nil
}

func (r *InMemoryRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	it, ok := r.items[key]
	if !ok {
		return 0, service.ErrNotFound
	}

	if it.expiresAt == 0 {
		return 0, nil
	}

	ttl := time.Duration(it.expiresAt-time.Now().Unix()) * time.Second

	if ttl < 0 {
		return 0, service.ErrNotFound
	}

	return ttl, nil
}

func (r *InMemoryRepository) Len(ctx context.Context) (int64, error) {
	return int64(len(r.items)), nil
}

func (r *InMemoryRepository) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				r.deleteExpired()
			case <-r.stopClean:
				ticker.Stop()
				return
			}

		}
	}()
}

func (r *InMemoryRepository) deleteExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix()
	for key, item := range r.items {
		if item.expiresAt > 0 && item.expiresAt < now {
			delete(r.items, key)
		}
	}
}

func (r *InMemoryRepository) StopCleanup() {
	close(r.stopClean)
}

func (r *InMemoryRepository) Clear(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = make(map[string]item)
	return nil
}
