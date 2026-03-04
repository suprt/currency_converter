package config

import (
	"os"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	t.Run("existing env", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		val := getEnv("TEST_VAR", "default")
		if val != "test_value" {
			t.Errorf("expected 'test_value', got '%s'", val)
		}
	})

	t.Run("non-existing env", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_VAR")

		val := getEnv("NON_EXISTENT_VAR", "default_value")
		if val != "default_value" {
			t.Errorf("expected 'default_value', got '%s'", val)
		}
	})

	t.Run("empty env", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		val := getEnv("EMPTY_VAR", "default_for_empty")
		if val != "default_for_empty" {
			t.Errorf("expected 'default_for_empty', got '%s'", val)
		}
	})
}

func TestGetEnvInt(t *testing.T) {
	t.Run("valid int", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		val := getEnvInt("TEST_INT", 0)
		if val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	t.Run("invalid int", func(t *testing.T) {
		os.Setenv("TEST_INT", "not_a_number")
		defer os.Unsetenv("TEST_INT")

		val := getEnvInt("TEST_INT", 100)
		if val != 100 {
			t.Errorf("expected 100 (default), got %d", val)
		}
	})

	t.Run("non-existing env", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_INT")

		val := getEnvInt("NON_EXISTENT_INT", 50)
		if val != 50 {
			t.Errorf("expected 50 (default), got %d", val)
		}
	})

	t.Run("empty env", func(t *testing.T) {
		os.Setenv("EMPTY_INT", "")
		defer os.Unsetenv("EMPTY_INT")

		val := getEnvInt("EMPTY_INT", 75)
		if val != 75 {
			t.Errorf("expected 75 (default), got %d", val)
		}
	})
}

func TestGetEnvDuration(t *testing.T) {
	t.Run("valid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "5m")
		defer os.Unsetenv("TEST_DURATION")

		val := getEnvDuration("TEST_DURATION", 0)
		if val != 5*time.Minute {
			t.Errorf("expected 5m, got %v", val)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "not_a_duration")
		defer os.Unsetenv("TEST_DURATION")

		val := getEnvDuration("TEST_DURATION", 10*time.Second)
		if val != 10*time.Second {
			t.Errorf("expected 10s (default), got %v", val)
		}
	})

	t.Run("non-existing env", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_DURATION")

		val := getEnvDuration("NON_EXISTENT_DURATION", 1*time.Hour)
		if val != 1*time.Hour {
			t.Errorf("expected 1h (default), got %v", val)
		}
	})

	t.Run("empty env", func(t *testing.T) {
		os.Setenv("EMPTY_DURATION", "")
		defer os.Unsetenv("EMPTY_DURATION")

		val := getEnvDuration("EMPTY_DURATION", 30*time.Second)
		if val != 30*time.Second {
			t.Errorf("expected 30s (default), got %v", val)
		}
	})
}

func TestGetEnvBool(t *testing.T) {
	t.Run("true value", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "true")
		defer os.Unsetenv("TEST_BOOL")

		val := getEnvBool("TEST_BOOL", false)
		if !val {
			t.Error("expected true, got false")
		}
	})

	t.Run("false value", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "false")
		defer os.Unsetenv("TEST_BOOL")

		val := getEnvBool("TEST_BOOL", true)
		if val {
			t.Error("expected false, got true")
		}
	})

	t.Run("invalid bool", func(t *testing.T) {
		os.Setenv("TEST_BOOL", "not_a_bool")
		defer os.Unsetenv("TEST_BOOL")

		val := getEnvBool("TEST_BOOL", true)
		if !val {
			t.Error("expected true (default), got false")
		}
	})

	t.Run("non-existing env", func(t *testing.T) {
		os.Unsetenv("NON_EXISTENT_BOOL")

		val := getEnvBool("NON_EXISTENT_BOOL", false)
		if val {
			t.Error("expected false (default), got true")
		}
	})

	t.Run("empty env", func(t *testing.T) {
		os.Setenv("EMPTY_BOOL", "")
		defer os.Unsetenv("EMPTY_BOOL")

		val := getEnvBool("EMPTY_BOOL", true)
		if !val {
			t.Error("expected true (default), got false")
		}
	})
}

func TestConfig_ServerAddr(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		expected string
	}{
		{
			name:     "localhost with port",
			host:     "localhost",
			port:     "8080",
			expected: "localhost:8080",
		},
		{
			name:     "empty host",
			host:     "",
			port:     ":8080",
			expected: "::8080", // host + ":" + port = "" + ":" + ":8080"
		},
		{
			name:     "IP address",
			host:     "0.0.0.0",
			port:     "3000",
			expected: "0.0.0.0:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ServerHost: tt.host,
				ServerPort: tt.port,
			}

			addr := cfg.ServerAddr()
			if addr != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, addr)
			}
		})
	}
}
