package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/suprt/currency_converter/internal/logger"
)

type ConverterClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type APIResponse struct {
	Valid   bool               `json:"valid"`
	Updated int64              `json:"updated"`
	Base    string             `json:"base"` //always USD
	Rates   map[string]float64 `json:"rates"`
}

type CurrenciesResponse struct {
	Valid      bool              `json:"valid"`
	Currencies map[string]string `json:"currencies"`
}

func NewConverterClient(baseURL string, apiKey string, timeout time.Duration) *ConverterClient {
	return &ConverterClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *ConverterClient) doGetRates(ctx context.Context) (rates map[string]float64, err error) {
	url := fmt.Sprintf("%srates?key=%s&base=USD&output=json", c.baseURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create get rates request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get rates request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response body failed: %w", err)
		}
		return nil, fmt.Errorf("get rates request failed, status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response body failed: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("API return invalid response")
	}

	return result.Rates, nil
}

func (c *ConverterClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var netError net.Error
	if errors.As(err, &netError) && netError.Timeout() {
		return true
	}
	if strings.Contains(err.Error(), "status code: 5") {
		return true
	}
	return false
}

func (c *ConverterClient) GetRates(ctx context.Context) (map[string]float64, error) {

	const (
		maxRetries = 3
		initDelay  = 1 * time.Second
	)

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := c.doGetRates(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err
		if logger.Log != nil {
			logger.Log.Warn("do get rates request", "attempt", attempt, "error", err)
		}
		if !c.isRetryableError(err) {
			return nil, err
		}

		delay := initDelay * time.Duration(1<<uint(attempt))

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

func (c *ConverterClient) doGetCurrencies(ctx context.Context) (body []byte, err error) {
	url := fmt.Sprintf("%scurrencies?key=%s&output=json", c.baseURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create get currencies request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get currencies request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response body failed: %w", err)
		}
		return nil, fmt.Errorf("get currencies request failed, status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	var result CurrenciesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response body failed: %w", err)
	}
	if !result.Valid {
		return nil, fmt.Errorf("API return invalid response")
	}

	return body, nil
}

func (c *ConverterClient) GetCurrencies(ctx context.Context) ([]byte, error) {

	const (
		maxRetries = 3
		initDelay  = 1 * time.Second
	)

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		body, err := c.doGetCurrencies(ctx)
		if err == nil {
			return body, nil
		}

		lastErr = err
		if logger.Log != nil {
			logger.Warn("do get currencies request", "attempt", attempt, "error", err)
		}
		if !c.isRetryableError(err) {
			return nil, err
		}

		delay := initDelay * time.Duration(1<<uint(attempt))

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}
