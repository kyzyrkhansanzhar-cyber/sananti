package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sananti/core"
)

type mockLogger struct {
	alerts []core.AlertData
}

func (m *mockLogger) LogAlert(alert core.AlertData) error {
	m.alerts = append(m.alerts, alert)
	return nil
}

func TestHoneyTrap_MiddlewareFlow(t *testing.T) {
	blocker := core.NewMemoryBlocker(1 * time.Hour)
	mlog := &mockLogger{}
	honeyTrap := NewHoneyTrap(blocker, mlog)

	// A basic mock target handler representing the legitimate application
	legitimateHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("LEGITIMATE"))
	})

	// Wrap our basic handler with the HoneyTrap interceptor
	trapURL := "/phpmyadmin"
	handler := honeyTrap.HandleTrap(trapURL)(legitimateHandler)

	// Wrap again with the global Protection Middleware
	handler = honeyTrap.ProtectionMiddleware()(handler)

	// --- Case 1: Standard legitimate request ---
	reqLegit := httptest.NewRequest("GET", "/dashboard", nil)
	reqLegit.RemoteAddr = "1.2.3.4:1111"
	wLegit := httptest.NewRecorder()

	handler.ServeHTTP(wLegit, reqLegit)

	if wLegit.Code != http.StatusOK {
		t.Errorf("expected status 200 for normal route, got %d", wLegit.Code)
	}
	if wLegit.Body.String() != "LEGITIMATE" {
		t.Errorf("expected response body 'LEGITIMATE', got %q", wLegit.Body.String())
	}
	if len(mlog.alerts) != 0 {
		t.Errorf("expected no intrusion alerts to be logged, got %d", len(mlog.alerts))
	}

	// --- Case 2: Scan request to decoy Honey URL ---
	attackerIP := "203.0.113.10"
	reqTrap := httptest.NewRequest("GET", trapURL, nil)
	reqTrap.RemoteAddr = attackerIP + ":2222"
	wTrap := httptest.NewRecorder()

	handler.ServeHTTP(wTrap, reqTrap)

	// Response code must be either 403 or 404 (randomly selected by design)
	if wTrap.Code != http.StatusForbidden && wTrap.Code != http.StatusNotFound {
		t.Errorf("expected trap response to return 403 or 404, got %d", wTrap.Code)
	}

	// Verify that the intrusion was registered
	if len(mlog.alerts) != 1 {
		t.Fatalf("expected exactly 1 alert, got %d", len(mlog.alerts))
	}
	if mlog.alerts[0].IP != attackerIP {
		t.Errorf("expected logged IP to match attacker IP %s, got %s", attackerIP, mlog.alerts[0].IP)
	}

	// Verify that the attacker IP was blacklisted
	blocked, reason, err := blocker.IsBlocked(attackerIP)
	if err != nil {
		t.Fatalf("IsBlocked lookup error: %v", err)
	}
	if !blocked {
		t.Errorf("expected attacker IP %s to be blocked in database", attackerIP)
	}
	if reason == "" {
		t.Error("expected blocking reason to not be empty")
	}

	// --- Case 3: Subsequent request to legitimate route by the same blocked IP ---
	reqBlocked := httptest.NewRequest("GET", "/dashboard", nil)
	reqBlocked.RemoteAddr = attackerIP + ":3333"
	wBlocked := httptest.NewRecorder()

	handler.ServeHTTP(wBlocked, reqBlocked)

	// Access must be denied instantly by the ProtectionMiddleware
	if wBlocked.Code != http.StatusForbidden {
		t.Errorf("expected subsequent request from blacklisted IP to return 403 Forbidden, got %d", wBlocked.Code)
	}
	if wBlocked.Body.String() == "LEGITIMATE" {
		t.Error("legitimate handler executed for blacklisted IP")
	}
}
