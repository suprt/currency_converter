package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

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
		t.Error("expected repo to be set")
	}
	if svc.client != client {
		t.Error("expected client to be set")
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
		{
			name:   "successful conversion",
			from:   "EUR",
			to:     "GBP",
			amount: 100,
			setupRates: map[string]float64{
				"USD:EUR": 0.85,
				"USD:GBP": 0.73,
			},
			expected:    85.88, // (1/0.85) * 0.73 * 100
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
				repo.data[key] = rate
			}

			client := NewMockConverterClient()
			svc := NewConverterService(repo, client)

			result, err := svc.Convert(context.Background(), tt.from, tt.to, tt.amount)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result == 0 && tt.expected != 0 {
				t.Errorf("expected non-zero result, got %f", result)
			}
		})
	}
}

func TestService_GetRates(t *testing.T) {
	t.Run("direct rate exists", func(t *testing.T) {
		repo := NewMockRepo()
		repo.data["EUR:GBP"] = 0.86

		client := NewMockConverterClient()
		svc := NewConverterService(repo, client)

		rate, err := svc.GetRates(context.Background(), "EUR", "GBP")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if rate != 0.86 {
			t.Errorf("expected rate 0.86, got %f", rate)
		}
	})

	t.Run("cross rate calculation", func(t *testing.T) {
		repo := NewMockRepo()
		repo.data["USD:EUR"] = 0.85
		repo.data["USD:GBP"] = 0.73

		client := NewMockConverterClient()
		svc := NewConverterService(repo, client)

		rate, err := svc.GetRates(context.Background(), "EUR", "GBP")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Expected: (1/0.85) * 0.73 = 0.8588...
		if rate < 0.85 || rate > 0.87 {
			t.Errorf("expected cross rate around 0.86, got %f", rate)
		}
	})

	t.Run("rates not found", func(t *testing.T) {
		repo := NewMockRepo()
		client := NewMockConverterClient()
		svc := NewConverterService(repo, client)

		_, err := svc.GetRates(context.Background(), "USD", "EUR")

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
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
			expected:    0.8588, // (1/0.85) * 0.73
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
				repo.data[key] = rate
			}

			client := NewMockConverterClient()
			svc := NewConverterService(repo, client)

			rate, err := svc.calculateCrossRate(context.Background(), tt.from, tt.to)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && err == nil {
				if rate < tt.expected*0.99 || rate > tt.expected*1.01 {
					t.Errorf("expected rate around %f, got %f", tt.expected, rate)
				}
			}
		})
	}
}

func TestService_RefreshRates(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		repo := NewMockRepo()
		client := NewMockConverterClient()
		client.SetRates(map[string]float64{
			"EUR": 0.85,
			"GBP": 0.73,
		})

		svc := NewConverterService(repo, client)

		err := svc.RefreshRates(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check if rates were set
		if len(repo.data) != 2 {
			t.Errorf("expected 2 rates, got %d", len(repo.data))
		}
	})

	t.Run("client error", func(t *testing.T) {
		repo := NewMockRepo()
		client := NewMockConverterClient()
		client.SetGetRatesError(assertionError("API error"))

		svc := NewConverterService(repo, client)

		err := svc.RefreshRates(context.Background())

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_Health(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		repo := NewMockRepo()
		repo.data["USD:USD"] = 1.0

		client := NewMockConverterClient()
		svc := NewConverterService(repo, client)

		err := svc.Health(context.Background())

		if err != nil {
			t.Errorf("expected healthy, got error: %v", err)
		}
	})

	t.Run("unhealthy - repo error", func(t *testing.T) {
		repo := NewMockRepo()
		repo.SetGetError(assertionError("storage error"))

		client := NewMockConverterClient()
		svc := NewConverterService(repo, client)

		err := svc.Health(context.Background())

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_GetLastUpdate(t *testing.T) {
	repo := NewMockRepo()
	client := NewMockConverterClient()
	svc := NewConverterService(repo, client)

	// Before refresh, last update should be zero (year 0000)
	lastUpdateBefore := svc.GetLastUpdate()
	if lastUpdateBefore.Year() != 0 {
		t.Errorf("expected zero time before refresh, got %v", lastUpdateBefore)
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
		t.Errorf("expected last update around now, got %v ago", diff)
	}
}

func TestService_Clear(t *testing.T) {
	repo := NewMockRepo()
	repo.data["USD:EUR"] = 0.85

	client := NewMockConverterClient()
	svc := NewConverterService(repo, client)

	err := svc.Clear(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(repo.data) != 0 {
		t.Error("expected repo to be cleared")
	}
}

func TestService_ForceRefresh(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		repo := NewMockRepo()
		repo.data["USD:OLD"] = 1.0

		client := NewMockConverterClient()
		client.SetRates(map[string]float64{
			"EUR": 0.85,
			"GBP": 0.73,
		})

		svc := NewConverterService(repo, client)

		err := svc.ForceRefresh(context.Background())

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Old key should be removed
		if _, exists := repo.data["USD:OLD"]; exists {
			t.Error("expected old data to be cleared")
		}
	})
}

// assertionError is a simple error type for testing
type assertionError string

func (e assertionError) Error() string { return string(e) }
