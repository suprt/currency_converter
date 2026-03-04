package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

func NewConverterClient(baseURL string, apiKey string) *ConverterClient {
	return &ConverterClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}

func (c *ConverterClient) GetRates(ctx context.Context) (map[string]float64, error) {
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

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

func (c *ConverterClient) GetCurrencies(ctx context.Context) ([]byte, error) {
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response body failed: %w", err)
		}
		return nil, fmt.Errorf("get currencies request failed, status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	return body, nil
}
