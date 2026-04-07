package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockCacheService - mock for CacheService interface
type MockCacheService struct {
	data            map[string]float64
	size            int64
	lastUpdate      time.Time
	getErr          error
	setErr          error
	deleteErr       error
	clearErr        error
	forceRefreshErr error
	ttlErr          error
	existsErr       error
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		data: make(map[string]float64),
	}
}

func (m *MockCacheService) DeleteRate(ctx context.Context, from, to string) error {
	return m.deleteErr
}

func (m *MockCacheService) ExistsRate(ctx context.Context, from, to string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	key := from + ":" + to
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockCacheService) CacheSize(ctx context.Context) (int64, error) {
	return m.size, nil
}

func (m *MockCacheService) GetKey(ctx context.Context, from, to string) (float64, error) {
	if m.getErr != nil {
		return 0, m.getErr
	}
	key := from + ":" + to
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return 0, assertionError("key not found")
}

func (m *MockCacheService) SetKey(ctx context.Context, from, to string, value float64, ttl time.Duration) error {
	return m.setErr
}

func (m *MockCacheService) GetLastUpdate() time.Time {
	return m.lastUpdate
}

func (m *MockCacheService) Clear(ctx context.Context) error {
	return m.clearErr
}

func (m *MockCacheService) ForceRefresh(ctx context.Context) error {
	return m.forceRefreshErr
}

func (m *MockCacheService) TTL(ctx context.Context, from, to string) (time.Duration, error) {
	if m.ttlErr != nil {
		return 0, m.ttlErr
	}
	return 1 * time.Hour, nil
}

func (m *MockCacheService) SetData(from, to string, value float64) {
	m.data[from+":"+to] = value
}

func TestCacheHandler_GetKey(t *testing.T) {

	tests := []struct {
		name           string
		from           string
		to             string
		value          float64
		getErr         error
		expectedStatus int
		expectedValue  float64
	}{
		{
			name:           "success",
			from:           "EUR",
			to:             "GBP",
			value:          0.86,
			expectedStatus: http.StatusOK,
			expectedValue:  0.86,
		},
		{
			name:           "missing parameters",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			from:           "EUR",
			to:             "GBP",
			getErr:         errors.New("cache miss"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockCacheService()
			svc.SetData(tt.from, tt.to, tt.value)
			svc.getErr = tt.getErr
			handler := NewCacheHandler(svc)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/cache/get?from=%s&to=%s", tt.from, tt.to), nil)
			w := httptest.NewRecorder()

			handler.GetKey(w, req)
			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]float64
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				key := tt.from + ":" + tt.to
				if val, ok := resp[key]; !ok || val != tt.expectedValue {
					t.Fatalf("expected value %f, got %f", tt.expectedValue, val)
				}
			}
		})
	}

}

func TestCacheHandler_SetKey(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		value          string
		ttl            string
		setErr         error
		expectedStatus int
	}{
		{"success", "EUR", "GBP", "0.86", "3600", nil, http.StatusOK},
		{"missing parameters", "", "", "", "", nil, http.StatusBadRequest},
		{"invalid ttl", "EUR", "GBP", "0.86", "-100", nil, http.StatusBadRequest},
		{"invalid value", "EUR", "GBP", "abc", "3600", nil, http.StatusBadRequest},
		{"service error", "EUR", "GBP", "0.86", "3600", errors.New("set failed"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockCacheService()
			svc.setErr = tt.setErr
			handler := NewCacheHandler(svc)

			req := httptest.NewRequest(http.MethodGet,
				fmt.Sprintf("/admin/cache/set?from=%s&to=%s&value=%s&ttl=%s", tt.from, tt.to, tt.value, tt.ttl),
				nil)
			w := httptest.NewRecorder()

			handler.SetKey(w, req)
			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

		})
	}

}

func TestCacheHandler_DeleteRate(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		deleteErr      error
		expectedStatus int
	}{
		{"success", "EUR", "GBP", nil, http.StatusOK},
		{"missing parameters", "", "", nil, http.StatusBadRequest},
		{"service error", "EUR", "GBP", errors.New("delete failed"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockCacheService()
			svc.deleteErr = tt.deleteErr
			handler := NewCacheHandler(svc)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/cache/delete?from=%s&to=%s", tt.from, tt.to), nil)
			w := httptest.NewRecorder()

			handler.DeleteRate(w, req)

			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}

}

func TestCacheHandler_CacheSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.size = 42
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/size", nil)
		w := httptest.NewRecorder()

		handler.CacheSize(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if size, ok := resp["size"].(float64); !ok || size != 42 {
			t.Fatalf("expected size 42, got %v", resp["size"])
		}
	})
}

func TestCacheHandler_CheckRate(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		value          float64
		expectedStatus int
		expectedExist  bool
	}{
		{"exists", "EUR", "GBP", 0.86, http.StatusOK, true},
		{"not exists", "EUR", "GBP", 0, http.StatusOK, false},
		{"missing parameters", "", "", 0, http.StatusBadRequest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockCacheService()
			if tt.expectedExist {
				svc.SetData(tt.from, tt.to, tt.value)
			}
			handler := NewCacheHandler(svc)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/cache/check?from=%s&to=%s", tt.from, tt.to), nil)
			w := httptest.NewRecorder()

			handler.CheckRate(w, req)
			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]bool
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if resp["exists"] != tt.expectedExist {
					t.Fatalf("expected value %t, got %t", tt.expectedExist, resp["exists"])
				}
			}
		})
	}
}

func TestCacheHandler_ClearAndRefresh(t *testing.T) {
	tests := []struct {
		name           string
		refreshErr     error
		expectedStatus int
	}{
		{"success", nil, http.StatusOK},
		{"clear error", errors.New("clear failed"), http.StatusInternalServerError},
		{"refresh error", errors.New("refresh failed"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockCacheService()
			svc.forceRefreshErr = tt.refreshErr
			handler := NewCacheHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/admin/cache/clear", nil)
			w := httptest.NewRecorder()

			handler.ClearAndRefresh(w, req)

			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}

}

func TestCacheHandler_TTLKey(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		ttlErr         error
		expectedStatus int
	}{
		{"success", "EUR", "GBP", nil, http.StatusOK},
		{"missing parameters", "", "", nil, http.StatusBadRequest},
		{"service error", "EUR", "GBP", errors.New("ttl error"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockCacheService()
			svc.ttlErr = tt.ttlErr
			handler := NewCacheHandler(svc)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/cache/ttl?from=%s&to=%s", tt.from, tt.to), nil)
			w := httptest.NewRecorder()

			handler.TTLKey(w, req)
			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if resp["from"] != tt.from || resp["to"] != tt.to {
					t.Fatalf("unexpected from/to in response")
				}
			}
		})
	}

}
