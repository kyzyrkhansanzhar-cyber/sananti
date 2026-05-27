package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sananti/core"
)

// MockLogger captures logged alerts for testing purposes.
type MockLogger struct {
	Alerts []core.AlertData
}

func (m *MockLogger) LogAlert(alert core.AlertData) error {
	m.Alerts = append(m.Alerts, alert)
	return nil
}

func TestRateLimiter_LimitMiddleware(t *testing.T) {
	blocker := core.NewMemoryBlocker(1 * time.Hour)
	logger := &MockLogger{}

	// Create a rate limiter: refill rate of 1 token per second, capacity/burst of 2
	limiter := NewRateLimiter(1.0, 2.0, blocker, logger)
	middleware := limiter.LimitMiddleware()

	// Simple mock handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	// 1. First request from client IP (consumes 1 token, remaining: 1.0)
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.1")
	resp1 := httptest.NewRecorder()
	handler.ServeHTTP(resp1, req1)
	if resp1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK on first request, got %d", resp1.Code)
	}

	// 2. Second request from same IP (consumes 1 token, remaining: 0.0)
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.1")
	resp2 := httptest.NewRecorder()
	handler.ServeHTTP(resp2, req2)
	if resp2.Code != http.StatusOK {
		t.Errorf("Expected 200 OK on second request, got %d", resp2.Code)
	}

	// 3. Third request instantly (0.0 tokens available -> should trigger 429)
	req3 := httptest.NewRequest("GET", "/api/data", nil)
	req3.Header.Set("X-Forwarded-For", "203.0.113.1")
	resp3 := httptest.NewRecorder()
	handler.ServeHTTP(resp3, req3)
	if resp3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests, got %d", resp3.Code)
	}

	// 4. Verify Blocker has blacklisted the client IP
	blocked, _, err := blocker.IsBlocked("203.0.113.1")
	if err != nil {
		t.Fatalf("Unexpected error from blocker: %v", err)
	}
	if !blocked {
		t.Error("Expected IP 203.0.113.1 to be blacklisted after rate limit breach")
	}

	// 5. Verify Logger captured the event
	if len(logger.Alerts) != 1 {
		t.Fatalf("Expected exactly 1 logged alert, got %d", len(logger.Alerts))
	}
	if logger.Alerts[0].IP != "203.0.113.1" || logger.Alerts[0].Severity != core.SeverityWarning {
		t.Errorf("Unexpected alert data recorded: %+v", logger.Alerts[0])
	}
}
