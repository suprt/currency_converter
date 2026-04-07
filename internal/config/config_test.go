package config

import (
	"os"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {

	tests := []struct {
		name       string
		key        string
		value      string
		expect     string
		defaultVal string
		keyExists  bool
	}{
		{
			name:       "existing key",
			key:        "TEST_VAR",
			value:      "test_value",
			expect:     "test_value",
			defaultVal: "default_value",
			keyExists:  true,
		},
		{
			name:       "non-existing key",
			key:        "NON_EXISTING_VAR",
			expect:     "default_value",
			defaultVal: "default_value",
			keyExists:  false,
		},
		{
			name:       "empty value",
			key:        "EMPTY_VAR",
			value:      "",
			expect:     "default_value",
			defaultVal: "default_value",
			keyExists:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv(tt.key)
			if tt.keyExists {
				_ = os.Setenv(tt.key, tt.defaultVal)
			}
			t.Cleanup(func() { _ = os.Unsetenv(tt.key) })

			val := getEnv(tt.key, tt.expect)
			if val != tt.expect {
				t.Fatalf("got %s, want %s", val, tt.expect)
			}
		})
	}

}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		expect     int
		defaultVal int
		keyExists  bool
	}{
		{
			name:       "valid int",
			key:        "TEST_INT",
			value:      "42",
			expect:     42,
			defaultVal: 0,
			keyExists:  true,
		},
		{
			name:       "invalid int",
			key:        "INVALID_INT",
			value:      "not_a_number",
			expect:     100,
			defaultVal: 100,
			keyExists:  true,
		},
		{
			name:       "non-existing int",
			key:        "NON_EXISTING_INT",
			expect:     50,
			defaultVal: 50,
			keyExists:  false,
		},
		{
			name:       "empty env",
			key:        "EMPTY_INT",
			value:      "",
			expect:     75,
			defaultVal: 75,
			keyExists:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv(tt.key)
			if tt.keyExists {
				_ = os.Setenv(tt.key, tt.value)
			}
			t.Cleanup(func() { _ = os.Unsetenv(tt.key) })
			val := getEnvInt(tt.key, tt.defaultVal)
			if val != tt.expect {
				t.Fatalf("got %d, want %d", val, tt.expect)
			}
		})
	}

}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		expect     time.Duration
		defaultVal time.Duration
		keyExists  bool
	}{
		{
			name:       "valid duration",
			key:        "VALID_DURATION",
			value:      "1m",
			expect:     1 * time.Minute,
			defaultVal: 0 * time.Minute,
			keyExists:  true,
		},
		{
			name:       "invalid duration",
			key:        "INVALID_DURATION",
			value:      "not_a_duration",
			expect:     10 * time.Second,
			defaultVal: 10 * time.Second,
			keyExists:  true,
		},
		{
			name:       "non-existing env",
			key:        "NON_EXISTING_DURATION",
			expect:     15 * time.Second,
			defaultVal: 15 * time.Second,
			keyExists:  false,
		},
		{
			name:       "empty env",
			key:        "EMPTY_DURATION",
			value:      "",
			expect:     30 * time.Second,
			defaultVal: 30 * time.Second,
			keyExists:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv(tt.key)
			if tt.keyExists {
				_ = os.Setenv(tt.key, tt.value)
			}
			t.Cleanup(func() { _ = os.Unsetenv(tt.key) })

			val := getEnvDuration(tt.key, tt.defaultVal)
			if val != tt.expect {
				t.Fatalf("got %s, want %s", val, tt.expect)
			}
		})
	}

}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		expect     bool
		defaultVal bool
		keyExists  bool
	}{
		{
			name:       "true value",
			key:        "TEST_BOOL",
			value:      "true",
			expect:     true,
			defaultVal: false,
			keyExists:  true,
		},
		{
			name:       "false value",
			key:        "TEST_BOOL",
			value:      "false",
			expect:     false,
			defaultVal: true,
			keyExists:  true,
		},
		{
			name:       "invalid bool",
			key:        "INVALID_BOOL",
			value:      "not_a_bool",
			expect:     true,
			defaultVal: true,
			keyExists:  true,
		},
		{
			name:       "non-existing env",
			key:        "NON_EXISTING_BOOL",
			expect:     false,
			defaultVal: false,
			keyExists:  false,
		},
		{
			name:       "empty env",
			key:        "EMPTY_BOOL",
			value:      "",
			expect:     true,
			defaultVal: true,
			keyExists:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv(tt.key)
			if tt.keyExists {
				_ = os.Setenv(tt.key, tt.value)
			}
			t.Cleanup(func() { _ = os.Unsetenv(tt.key) })

			val := getEnvBool(tt.key, tt.defaultVal)
			if val != tt.expect {
				t.Fatalf("got %t, want %t", val, tt.expect)
			}
		})
	}
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
				t.Fatalf("expected '%s', got '%s'", tt.expected, addr)
			}
		})
	}
}
