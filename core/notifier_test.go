package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookNotifier_SendAlert_Success(t *testing.T) {
	var receivedAlert AlertData
	var handlerCalled bool

	// Spin up a mock HTTP server to receive the webhook alert
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json Content-Type, got %s", r.Header.Get("Content-Type"))
		}

		err := json.NewDecoder(r.Body).Decode(&receivedAlert)
		if err != nil {
			t.Errorf("Failed to decode webhook payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	notifier := NewWebhookNotifier(ts.URL, 1*time.Second)

	alert := AlertData{
		IP:        "1.2.3.4",
		Path:      "/test-webhook",
		Method:    "GET",
		Severity:  SeverityCritical,
		Timestamp: time.Now(),
		Details:   "Testing critical webhook dispatcher",
	}

	err := notifier.SendAlert(alert)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !handlerCalled {
		t.Fatal("Mock server handler was not called")
	}

	if receivedAlert.IP != "1.2.3.4" || receivedAlert.Details != "Testing critical webhook dispatcher" {
		t.Errorf("Received mismatch payload: %+v", receivedAlert)
	}
}

func TestWebhookNotifier_SendAlert_FailResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	notifier := NewWebhookNotifier(ts.URL, 1*time.Second)

	alert := AlertData{
		IP:        "9.9.9.9",
		Path:      "/error-hook",
		Severity:  SeverityWarning,
		Timestamp: time.Now(),
		Details:   "Testing fail response",
	}

	err := notifier.SendAlert(alert)
	if err == nil {
		t.Fatal("Expected error due to 500 Internal Server Error, got nil")
	}
}
