package service

import (
	"context"
	"sync"
	"time"
)

// MockRepo - mock implementation of Repo interface
type MockRepo struct {
	mu       sync.RWMutex
	data     map[string]float64
	ttlData  map[string]time.Time
	getError error
	setError error
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		data:    make(map[string]float64),
		ttlData: make(map[string]time.Time),
	}
}

func (m *MockRepo) Get(ctx context.Context, key string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getError != nil {
		return 0, m.getError
	}

	val, ok := m.data[key]
	if !ok {
		return 0, ErrNotFound
	}

	// Check TTL
	if ttl, ok := m.ttlData[key]; ok && time.Now().After(ttl) {
		return 0, ErrNotFound
	}

	return val, nil
}

func (m *MockRepo) Set(ctx context.Context, key string, value float64, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.setError != nil {
		return m.setError
	}

	m.data[key] = value
	if ttl > 0 {
		m.ttlData[key] = time.Now().Add(ttl)
	}
	return nil
}

func (m *MockRepo) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	delete(m.ttlData, key)
	return nil
}

func (m *MockRepo) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}

func (m *MockRepo) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.data[key]; !ok {
		return 0, ErrNotFound
	}

	if ttl, ok := m.ttlData[key]; ok {
		return time.Until(ttl), nil
	}
	return 0, nil
}

func (m *MockRepo) Len(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.data)), nil
}

func (m *MockRepo) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]float64)
	m.ttlData = make(map[string]time.Time)
	return nil
}

// SetGetError sets error to return from Get method
func (m *MockRepo) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getError = err
}

// SetSetError sets error to return from Set method
func (m *MockRepo) SetSetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setError = err
}

// MockConverterClient - mock implementation of ConverterClient interface
type MockConverterClient struct {
	rates           map[string]float64
	currencies      []byte
	getRatesError   error
	getCurrenciesErr error
}

func NewMockConverterClient() *MockConverterClient {
	return &MockConverterClient{
		rates: make(map[string]float64),
	}
}

func (m *MockConverterClient) GetRates(ctx context.Context) (map[string]float64, error) {
	if m.getRatesError != nil {
		return nil, m.getRatesError
	}
	return m.rates, nil
}

func (m *MockConverterClient) GetCurrencies(ctx context.Context) ([]byte, error) {
	if m.getCurrenciesErr != nil {
		return nil, m.getCurrenciesErr
	}
	return m.currencies, nil
}

func (m *MockConverterClient) SetRates(rates map[string]float64) {
	m.rates = rates
}

func (m *MockConverterClient) SetCurrencies(data []byte) {
	m.currencies = data
}

func (m *MockConverterClient) SetGetRatesError(err error) {
	m.getRatesError = err
}

func (m *MockConverterClient) SetGetCurrenciesError(err error) {
	m.getCurrenciesErr = err
}
