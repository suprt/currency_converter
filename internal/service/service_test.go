package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/suprt/currency_converter/internal/logger"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	os.Exit(m.Run())
}

func TestNewConverterService(t *testing.T) {
	repo := NewMockRepo()
	client := NewMockConverterClient()

	svc := NewConverterService(repo, client)

	if svc == nil {
		t.Fatal("expected service to be created")
	}
	if svc.repo != repo {
		t.Fatal("expected repo to be set")
	}
	if svc.client != client {
		t.Fatal("expected client to be set")
	}
}

func TestService_Convert(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		amount      float64
		setupRates  map[string]float64
		expected    float64
		expectError bool
	}{
		{name: "successful conversion",
			from:   "EUR",
			to:     "GBP",
			amount: 100,
			setupRates: map[string]float64{
				"USD:EUR": 0.85,
				"USD:GBP": 0.73,
			},
			expected:    (1 / 0.85) * 0.73 * 100,
			expectError: false,
		},
		{
			name:   "zero amount",
			from:   "USD",
			to:     "EUR",
			amount: 0,
			setupRates: map[string]float64{
				"USD:EUR": 0.85,
			},
			expected:    0,
			expectError: false,
		},
		{
			name:        "missing rates",
			from:        "USD",
			to:          "EUR",
			amount:      100,
			setupRates:  map[string]float64{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepo()
			for key, rate := range tt.setupRates {
				_ = repo.Set(context.Background(), key, rate, 0)
			}

			client := NewMockConverterClient()
			svc := NewConverterService(repo, client)

			result, err := svc.Convert(context.Background(), tt.from, tt.to, tt.amount)

			if tt.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.expectError && result == 0 && tt.expected != 0 {
				t.Fatalf("expected non-zero result, got %f", result)
			}
			if !tt.expectError && !assert.InDelta(t, tt.expected, result, 0.001) {
				t.Fatalf("expected %f, got %f", tt.expected, result)
			}
		})

	}
}

func TestService_GetRates(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		setupRates  map[string]float64
		expected    float64
		expectError bool
	}{
		{
			name:        "direct rate exists",
			from:        "EUR",
			to:          "GBP",
			setupRates:  map[string]float64{"EUR:GBP": 0.86},
			expected:    0.86,
			expectError: false,
		},
		{
			name:        "cross rate calculation",
			from:        "EUR",
			to:          "GBP",
			setupRates:  map[string]float64{"USD:EUR": 0.85, "USD:GBP": 0.73},
			expected:    1 / 0.85 * 0.73,
			expectError: false,
		},
		{
			name:        "rates not found",
			from:        "USD",
			to:          "EUR",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepo()
			for key, rate := range tt.setupRates {
				_ = repo.Set(context.Background(), key, rate, 0)
			}

			client := NewMockConverterClient()
			svc := NewConverterService(repo, client)

			result, err := svc.GetRates(context.Background(), tt.from, tt.to)

			if tt.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.expectError && !assert.InDelta(t, tt.expected, result, 0.001) {
				t.Fatalf("expected %f, got %f", tt.expected, result)
			}

		})

	}
}

func TestService_calculateCrossRate(t *testing.T) {
	tests := []struct {
		name        string
		setupRates  map[string]float64
		from        string
		to          string
		expected    float64
		expectError bool
	}{
		{
			name: "successful cross rate",
			setupRates: map[string]float64{
				"USD:EUR": 0.85,
				"USD:GBP": 0.73,
			},
			from:        "EUR",
			to:          "GBP",
			expected:    (1 / 0.85) * 0.73,
			expectError: false,
		},
		{
			name:        "missing from rate",
			setupRates:  map[string]float64{"USD:GBP": 0.73},
			from:        "EUR",
			to:          "GBP",
			expectError: true,
		},
		{
			name:        "missing to rate",
			setupRates:  map[string]float64{"USD:EUR": 0.85},
			from:        "EUR",
			to:          "GBP",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepo()
			for key, rate := range tt.setupRates {
				_ = repo.Set(context.Background(), key, rate, 0)
			}

			client := NewMockConverterClient()
			svc := NewConverterService(repo, client)

			rate, err := svc.calculateCrossRate(context.Background(), tt.from, tt.to)

			if tt.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.expectError && !assert.InDelta(t, tt.expected, rate, 0.001) {
				t.Fatalf("expected %f, got %f", tt.expected, rate)
			}

		})
	}
}

func TestService_RefreshRates(t *testing.T) {
	tests := []struct {
		name        string
		setupRates  map[string]float64
		expectError bool
	}{
		{
			name: "successful refresh",
			setupRates: map[string]float64{
				"EUR": 0.85,
				"GBP": 0.73,
			},
			expectError: false,
		},
		{
			name:        "client error",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepo()
			client := NewMockConverterClient()
			client.SetRates(tt.setupRates)
			svc := NewConverterService(repo, client)
			if tt.expectError {
				client.SetGetRatesError(errors.New("API error"))
			}
			err := svc.RefreshRates(context.Background())

			if tt.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rLen, err := repo.Len(context.Background()); err != nil {
				if !tt.expectError {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if (len(tt.setupRates)) != int(rLen) {
				t.Fatalf("expected %d rates, got %d", len(tt.setupRates), rLen)
			}
		})
	}
}

func TestService_Health(t *testing.T) {
	tests := []struct {
		name        string
		setupRates  map[string]float64
		expectError bool
	}{
		{
			name: "healthy",
			setupRates: map[string]float64{
				"USD:EUR": 0.85,
			},
			expectError: false,
		},
		{
			name:        "unhealthy - repo error",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepo()
			for key, rate := range tt.setupRates {
				_ = repo.Set(context.Background(), key, rate, 0)
			}
			if tt.expectError {
				repo.SetGetError(errors.New("storage error"))
			}
			client := NewMockConverterClient()
			svc := NewConverterService(repo, client)

			err := svc.Health(context.Background())
			if tt.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_GetLastUpdate(t *testing.T) {
	repo := NewMockRepo()
	client := NewMockConverterClient()
	svc := NewConverterService(repo, client)

	// Before refresh, last update should be zero (year 0000)
	lastUpdateBefore := svc.GetLastUpdate()
	if lastUpdateBefore.Year() != 0 {
		t.Fatalf("expected zero time before refresh, got %v", lastUpdateBefore)
	}

	// Set nextRefreshTime by refreshing
	client.SetRates(map[string]float64{"EUR": 0.85})
	err := svc.RefreshRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error during refresh: %v", err)
	}

	lastUpdate := svc.GetLastUpdate()

	// Last update should be approximately now (within 1 minute)
	diff := time.Since(lastUpdate)
	if diff < 0 || diff > 1*time.Minute {
		t.Fatalf("expected last update around now, got %v ago", diff)
	}
}

func TestService_Clear(t *testing.T) {
	repo := NewMockRepo()
	_ = repo.Set(context.Background(), "USD:EUR", 0.85, 0)
	client := NewMockConverterClient()
	svc := NewConverterService(repo, client)

	err := svc.Clear(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rLen, err := repo.Len(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)

	} else if rLen != 0 {
		t.Fatalf("expected repo to be cleared")
	}
}

func TestService_ForceRefresh(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		repo := NewMockRepo()
		_ = repo.Set(context.Background(), "USD:USD", 1.0, 0)
		client := NewMockConverterClient()
		client.SetRates(map[string]float64{"EUR": 0.85, "GBP": 0.7})
		svc := NewConverterService(repo, client)

		err := svc.ForceRefresh(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := repo.Get(context.Background(), "USD:USD"); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
		if val, err := repo.Get(context.Background(), "USD:EUR"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if val != 0.85 {
			t.Fatalf("expected 0.85, got %f", val)
		}
	})
}
