package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockCacheService - mock for CacheService interface
type MockCacheService struct {
	data        map[string]float64
	size        int64
	lastUpdate  time.Time
	getErr      error
	setErr      error
	deleteErr   error
	clearErr    error
	forceRefreshErr error
	ttlErr      error
	existsErr   error
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
	t.Run("success", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.SetData("EUR", "GBP", 0.86)
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/get?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.GetKey(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]float64
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		key := "EUR:GBP"
		if val, ok := resp[key]; !ok || val != 0.86 {
			t.Errorf("expected %s=0.86, got %v", key, resp)
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/get", nil)
		w := httptest.NewRecorder()

		handler.GetKey(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.getErr = assertionError("cache miss")
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/get?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.GetKey(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestCacheHandler_SetKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/set?from=EUR&to=GBP&value=0.86&ttl=3600", nil)
		w := httptest.NewRecorder()

		handler.SetKey(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/set", nil)
		w := httptest.NewRecorder()

		handler.SetKey(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid ttl", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/set?from=EUR&to=GBP&value=0.86&ttl=-100", nil)
		w := httptest.NewRecorder()

		handler.SetKey(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/set?from=EUR&to=GBP&value=invalid&ttl=3600", nil)
		w := httptest.NewRecorder()

		handler.SetKey(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCacheHandler_DeleteRate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/delete?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.DeleteRate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/delete", nil)
		w := httptest.NewRecorder()

		handler.DeleteRate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.deleteErr = assertionError("delete failed")
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/delete?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.DeleteRate(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
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
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if size, ok := resp["size"].(float64); !ok || size != 42 {
			t.Errorf("expected size 42, got %v", resp["size"])
		}
	})
}

func TestCacheHandler_CheckRate(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.SetData("EUR", "GBP", 0.86)
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/check?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.CheckRate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]bool
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if !resp["exists"] {
			t.Error("expected key to exist")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/check?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.CheckRate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]bool
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["exists"] {
			t.Error("expected key to not exist")
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/check", nil)
		w := httptest.NewRecorder()

		handler.CheckRate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestCacheHandler_ClearAndRefresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/admin/cache/clear", nil)
		w := httptest.NewRecorder()

		handler.ClearAndRefresh(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("clear error", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.clearErr = assertionError("clear failed")
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/admin/cache/clear", nil)
		w := httptest.NewRecorder()

		handler.ClearAndRefresh(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("refresh error", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.forceRefreshErr = assertionError("refresh failed")
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/admin/cache/clear", nil)
		w := httptest.NewRecorder()

		handler.ClearAndRefresh(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestCacheHandler_TTLKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/ttl?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.TTLKey(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["from"] != "EUR" || resp["to"] != "GBP" {
			t.Error("unexpected from/to in response")
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		svc := NewMockCacheService()
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/ttl", nil)
		w := httptest.NewRecorder()

		handler.TTLKey(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := NewMockCacheService()
		svc.ttlErr = assertionError("ttl error")
		handler := NewCacheHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/admin/cache/ttl?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.TTLKey(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}
