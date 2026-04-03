package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ===== GetRates Tests =====
func TestConverterClient_GetRates_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid": true, "updated": 1234567890, "base":"USD", "rates":{"EUR":0.85, "GBP":0.73}}`))
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)
	rates, err := client.GetRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if rates["EUR"] != 0.85 {
		t.Errorf("EUR rate is %f, expected %f", rates["EUR"], 0.85)
	}
	if rates["GBP"] != 0.73 {
		t.Errorf("GBP rate is %f, expected %f", rates["GBP"], 0.73)
	}
}

func TestConverterClient_GetRates_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"not valid json":json}`))
	}))
	defer server.Close()
	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)
	_, err := client.GetRates(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "parse response body failed") {
		t.Fatalf("unexpected error: %s", err)
	}

}

func TestConverterClient_GetRates_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

	}))
	defer server.Close()
	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)
	_, err := client.GetRates(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("unexpected error: %s", err)
	}

}

func TestConverterClient_GetRates_NetworkError(t *testing.T) {
	client := NewConverterClient("https://InvalidHostName/", "test-api-key", 15*time.Second, nil)
	_, err := client.GetRates(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "dial tcp") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestConverterClient_GetRates_ValidFalseResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid": false, "updated": 1234567890, "base":"USD", "rates":{"EUR":0.85, "GBP":0.73}}`))
	}))
	defer server.Close()
	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)
	_, err := client.GetRates(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "API return invalid response") {
		t.Fatalf("unexpected error: %s", err)
	}

}

func TestConverterClient_GetRates_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", time.Second, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := client.GetRates(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestConverterClient_GetRates_RetryOn500(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"valid": true, "updated": 1234567890, "base":"USD", "rates":{"EUR":0.85, "GBP":0.73}}`))
		}
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)

	rates, err := client.GetRates(context.Background())

	if err != nil {
		t.Fatalf("expected success after retries, got error: %s", err)
	}
	if rates["EUR"] != 0.85 {
		t.Errorf("EUR rate is %f, expected %f", rates["EUR"], 0.85)
	}
	if rates["GBP"] != 0.73 {
		t.Errorf("GBP rate is %f, expected %f", rates["GBP"], 0.73)
	}
}

func TestConverterClient_GetRates_RetryOnTimeout(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		attempt++
		if attempt < 3 {
			time.Sleep(2 * time.Second)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"valid": true, "updated": 1234567890, "base":"USD", "rates":{"EUR":0.85, "GBP":0.73}}`))
		}
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 1*time.Second, nil)

	rates, err := client.GetRates(context.Background())

	if err != nil {
		t.Fatalf("expected success after retries, got error: %s", err)
	}
	if rates["EUR"] != 0.85 {
		t.Errorf("EUR rate is %f, expected %f", rates["EUR"], 0.85)
	}
	if rates["GBP"] != 0.73 {
		t.Errorf("GBP rate is %f, expected %f", rates["GBP"], 0.73)
	}
}

func TestConverterClient_GetRates_400(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"valid": true, "updated": 1234567890, "base":"USD", "rates":{"EUR":0.85, "GBP":0.73}}`))
		}
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 1*time.Second, nil)

	_, err := client.GetRates(context.Background())
	if attempt > 1 {
		t.Fatalf("expected %v attempt, do %v", 1, attempt)
	}
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "4") {
		t.Fatalf("expected 4xx error, got %s", err)
	}
}

// ===== GetCurrencies Tests =====
func TestConverterClient_GetCurrencies_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid": true, "currencies":{"USD":"US Dollar", "EUR":"Euro"}}`))
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)
	currencies, err := client.GetCurrencies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(currencies) == 0 {
		t.Fatalf("expected non-empty response")
	}
}

func TestConverterClient_GetCurrencies_NetworkError(t *testing.T) {
	client := NewConverterClient("https://InvalidHostName/", "test-api-key", 15*time.Second, nil)
	_, err := client.GetCurrencies(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "dial tcp") {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestConverterClient_GetCurrencies_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

	}))
	defer server.Close()
	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)
	_, err := client.GetCurrencies(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestConverterClient_GetCurrencies_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()
	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)
	_, err := client.GetCurrencies(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "parse response body failed") {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestConverterClient_GetCurrencies_RetryOn500(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"valid": true, "currencies":{"USD":"US Dollar", "EUR":"Euro"}}`))
		}
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 15*time.Second, nil)

	currencies, err := client.GetCurrencies(context.Background())

	if err != nil {
		t.Fatalf("expected success after retries, got error: %s", err)
	}
	if len(currencies) == 0 {
		t.Fatalf("expected non-empty response")
	}
}

func TestConverterClient_GetCurrencies_RetryOnTimeout(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		attempt++
		if attempt < 3 {
			time.Sleep(2 * time.Second)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"valid": true, "currencies":{"USD":"US Dollar", "EUR":"Euro"}}`))
		}
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 1*time.Second, nil)

	currencies, err := client.GetCurrencies(context.Background())

	if err != nil {
		t.Fatalf("expected success after retries, got error: %s", err)
	}
	if len(currencies) == 0 {
		t.Fatalf("expected non-empty response")
	}
}

func TestConverterClient_GetCurrencies_400(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"valid": true, "currencies":{"USD":"US Dollar", "EUR":"Euro"}}`))
		}
	}))
	defer server.Close()

	client := NewConverterClient(server.URL+"/", "test-api-key", 1*time.Second, nil)

	_, err := client.GetCurrencies(context.Background())
	if attempt > 1 {
		t.Fatalf("expected exactly %d attempt got %d", 1, attempt)
	}
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "4") {
		t.Fatalf("expected 4xx error, got %s", err)
	}
}
