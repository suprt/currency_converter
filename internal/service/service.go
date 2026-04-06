package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/suprt/currency_converter/internal/logger"
)

type Repo interface {
	Get(ctx context.Context, key string) (float64, error)
	Set(ctx context.Context, key string, value float64, ttl time.Duration) error
	Delete(ctx context.Context, key string) error

	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Len(ctx context.Context) (int64, error)
	Clear(ctx context.Context) error
}

type ConverterClient interface {
	GetRates(ctx context.Context) (map[string]float64, error)
	GetCurrencies(ctx context.Context) ([]byte, error)
}

var ErrNotFound = errors.New("key not found")

type Service struct {
	repo   Repo
	client ConverterClient

	nextRefreshTime time.Time
}

func (s *Service) Convert(ctx context.Context, from string, to string, amount float64) (float64, error) {
	rate, err := s.GetRates(ctx, from, to)
	if err != nil {
		return 0, err
	}
	result := amount * rate
	return result, nil
}

func NewConverterService(repo Repo, client ConverterClient) *Service {
	return &Service{repo: repo, client: client}
}

func (s *Service) RefreshRates(ctx context.Context) error {
	logger.Info("refreshing currencies rates")

	resp, err := s.client.GetRates(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch currencies rates: %w", err)
	}
	if err := s.repo.Clear(ctx); err != nil {
		logger.Error("failed to clear currencies rates", "error", err)
	}

	for currency, rate := range resp {
		key := fmt.Sprintf("%s:%s", "USD", currency)
		if err := s.repo.Set(ctx, key, rate, time.Hour); err != nil {
			logger.Error("failed to cache rate", "key", key, "error", err)
		}
	}
	s.nextRefreshTime = time.Now().Add(time.Hour)
	logger.Info("successfully refreshed currencies rates")
	return nil
}

func (s *Service) GetRates(ctx context.Context, from, to string) (float64, error) {
	directKey := fmt.Sprintf("%s:%s", from, to)
	if rate, err := s.repo.Get(ctx, directKey); err == nil {
		return rate, nil
	}
	rate, err := s.calculateCrossRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	ttl := time.Until(s.nextRefreshTime)
	if ttl > 0 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in cache rate goroutine", "key", directKey, "panic", r)
				}
			}()
			select {
			case <-ctx.Done():
				return
			default:
			}
			err := s.repo.Set(ctx, directKey, rate, ttl)
			if err != nil {
				logger.Error("failed to cache rate", "key", directKey, "error", err)
			}
		}()
	}
	return rate, nil
}

func (s *Service) calculateCrossRate(ctx context.Context, from string, to string) (float64, error) {
	fromKey := fmt.Sprintf("%s:%s", "USD", from)
	rateFromUSD, err := s.repo.Get(ctx, fromKey)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch rate for %s: %w", from, err)
	}
	toKey := fmt.Sprintf("%s:%s", "USD", to)
	rateToUSD, err := s.repo.Get(ctx, toKey)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch rate for %s: %w", to, err)
	}
	return (1 / rateFromUSD) * rateToUSD, nil
}

func (s *Service) GetCurrenciesJSON(ctx context.Context) ([]byte, error) {
	return s.client.GetCurrencies(ctx)
}

func (s *Service) GetLastUpdate() time.Time {
	return s.nextRefreshTime.Add(-time.Hour)
}
func (s *Service) Health(ctx context.Context) error {
	_, err := s.repo.Get(ctx, "USD:USD")
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("storage check failed: %w", err)
	}
	return nil
}

func (s *Service) DeleteRate(ctx context.Context, from, to string) error {
	key := fmt.Sprintf("%s:%s", from, to)
	return s.repo.Delete(ctx, key)
}
func (s *Service) ExistsRate(ctx context.Context, from, to string) (bool, error) {
	key := fmt.Sprintf("%s:%s", from, to)
	return s.repo.Exists(ctx, key)
}

func (s *Service) TTL(ctx context.Context, from, to string) (time.Duration, error) {
	key := fmt.Sprintf("%s:%s", from, to)
	return s.repo.TTL(ctx, key)
}

// Прямой доступ к ключу
func (s *Service) GetKey(ctx context.Context, from, to string) (float64, error) {
	key := fmt.Sprintf("%s:%s", from, to)
	return s.repo.Get(ctx, key)
}

// Прямая установка ключа
func (s *Service) SetKey(ctx context.Context, from, to string, value float64, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", from, to)
	return s.repo.Set(ctx, key, value, ttl)
}

func (s *Service) CacheSize(ctx context.Context) (int64, error) {
	return s.repo.Len(ctx)
}

// Clear Принудительная очистка кэша
func (s *Service) Clear(ctx context.Context) error {
	return s.repo.Clear(ctx)
}

// ForceRefresh Принудительное обновление кэша
func (s *Service) ForceRefresh(ctx context.Context) error {
	resp, err := s.client.GetRates(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch currencies rates: %w", err)
	}
	if err := s.repo.Clear(ctx); err != nil {
		logger.Error("failed to clear currencies rates", "error", err)
	}

	for currency, rate := range resp {
		key := fmt.Sprintf("%s:%s", "USD", currency)
		if err := s.repo.Set(ctx, key, rate, time.Hour); err != nil {
			logger.Error("failed to cache rate", "key", key, "error", err)
		}
	}
	return nil
}
