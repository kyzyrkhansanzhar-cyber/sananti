package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sananti/antifraud"
	"sananti/core"
)

func TestAntiFraudMiddleware_AutoScan(t *testing.T) {
	blocker := core.NewMemoryBlocker(1 * time.Hour)
	logger, _ := core.NewFileLogger("logs/test_antifraud.log", 100, 1024*1024)
	defer blocker.Close()
	defer logger.Close(context.Background())

	// Initialize scanner with RecipientBlacklistRule
	rules := []antifraud.Rule{
		antifraud.NewRecipientBlacklistRule(),
	}
	scanner := antifraud.NewAntiFraudScanner(blocker, logger, rules...)

	// Create our zero-touch background middleware
	afm := NewAntiFraudMiddleware(scanner, blocker, logger)
	middleware := afm.AutoScanMiddleware()

	// Simple endpoint handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PAYMENT_PROCESSED_SUCCESSFULLY"))
	}))

	// --- Case 1: Send money to a SAFE recipient ---
	safePayload := map[string]interface{}{
		"user_id":          "user_1",
		"recipient_phone":  "+7 707 111 22 33", // Safe phone
		"amount":           100.0,
		"ip":               "127.0.0.1",
	}
	safeData, _ := json.Marshal(safePayload)

	reqSafe := httptest.NewRequest("POST", "/api/v1/payment", bytes.NewBuffer(safeData))
	reqSafe.Header.Set("Content-Type", "application/json")
	respSafe := httptest.NewRecorder()

	handler.ServeHTTP(respSafe, reqSafe)
	if respSafe.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for safe recipient, got %d", respSafe.Code)
	}

	// --- Case 2: Send money to a BLACKLISTED SCAMMER recipient ---
	scamPayload := map[string]interface{}{
		"user_id":          "user_1",
		"recipient_phone":  "+77777777777", // Blacklisted scam number!
		"amount":           100.0,
		"ip":               "127.0.0.1",
	}
	scamData, _ := json.Marshal(scamPayload)

	reqScam := httptest.NewRequest("POST", "/api/v1/payment", bytes.NewBuffer(scamData))
	reqScam.Header.Set("Content-Type", "application/json")
	respScam := httptest.NewRecorder()

	handler.ServeHTTP(respScam, reqScam)
	if respScam.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for scammer recipient, got %d", respScam.Code)
	}

	// Decode response to verify JSON content
	var body map[string]interface{}
	_ = json.Unmarshal(respScam.Body.Bytes(), &body)
	if body["status"] != "blocked" || body["error"] != "FRAUD_DETECTED" {
		t.Errorf("Unexpected blocked payload returned: %+v", body)
	}
}
