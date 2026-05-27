package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"sananti/core"
)

func TestHoneyField_BotInterception(t *testing.T) {
	blocker := core.NewMemoryBlocker(1 * time.Hour)
	defer blocker.Close()
	mlog := &mockLogger{} // defined in honeytrap_test.go

	fieldName := "subscribe_newsletter_optin"
	// Set 500ms time-lock threshold
	honeyField := NewHoneyField(blocker, mlog, fieldName, 500*time.Millisecond)

	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SUCCESS"))
	})

	handler := honeyField.HandleField()(targetHandler)

	// --- Case 1: Legitimate Form Submission (Honeypot empty & Time-lock satisfied) ---
	token := honeyField.GenerateTimeLockToken()
	
	// Simulate human waiting 600ms before submitting
	time.Sleep(600 * time.Millisecond)

	formSafe := url.Values{}
	formSafe.Set("name", "John Doe")
	formSafe.Set("email", "john@example.com")
	formSafe.Set(fieldName, "") // Hidden field left empty
	formSafe.Set("_sananti_timelock", token)

	reqSafe := httptest.NewRequest("POST", "/contact", strings.NewReader(formSafe.Encode()))
	reqSafe.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqSafe.RemoteAddr = "192.168.1.10:4567"
	wSafe := httptest.NewRecorder()

	handler.ServeHTTP(wSafe, reqSafe)

	if wSafe.Code != http.StatusOK {
		t.Errorf("expected 200 OK for safe form submission, got %d", wSafe.Code)
	}
	if wSafe.Body.String() != "SUCCESS" {
		t.Errorf("expected response body 'SUCCESS', got %q", wSafe.Body.String())
	}

	// --- Case 2: Bot Form Submission (Honeypot field filled by bot) ---
	botIP := "198.51.100.5"
	tokenBot := honeyField.GenerateTimeLockToken()
	time.Sleep(600 * time.Millisecond) // Wait to bypass speed lock, but trigger honey field

	formBot := url.Values{}
	formBot.Set("name", "Spam Bot")
	formBot.Set("email", "spam@bot.com")
	formBot.Set(fieldName, "subscribeme") // Decoy field filled by scraper
	formBot.Set("_sananti_timelock", tokenBot)

	reqBot := httptest.NewRequest("POST", "/contact", strings.NewReader(formBot.Encode()))
	reqBot.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqBot.RemoteAddr = botIP + ":8888"
	wBot := httptest.NewRecorder()

	handler.ServeHTTP(wBot, reqBot)

	if wBot.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for bot submission, got %d", wBot.Code)
	}

	// Verify that the bot IP was blacklisted
	blocked, _, _ := blocker.IsBlocked(botIP)
	if !blocked {
		t.Errorf("expected bot IP %s to be blacklisted in core blocker", botIP)
	}

	// --- Case 3: Bot Form Submission (Speed-Lock Triggered) ---
	fastIP := "198.51.100.6"
	tokenFast := honeyField.GenerateTimeLockToken() // Generate token
	// SUBMIT IMMEDIATELY (0ms sleep) to trigger the time-lock

	formFast := url.Values{}
	formFast.Set("name", "Fast Bot")
	formFast.Set("email", "fast@bot.com")
	formFast.Set(fieldName, "") // Hidden field is clean
	formFast.Set("_sananti_timelock", tokenFast)

	reqFast := httptest.NewRequest("POST", "/contact", strings.NewReader(formFast.Encode()))
	reqFast.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqFast.RemoteAddr = fastIP + ":9999"
	wFast := httptest.NewRecorder()

	handler.ServeHTTP(wFast, reqFast)

	if wFast.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden due to speed-lock, got %d", wFast.Code)
	}

	// Verify that the fast IP was blacklisted
	blockedFast, _, _ := blocker.IsBlocked(fastIP)
	if !blockedFast {
		t.Errorf("expected fast IP %s to be blacklisted due to speed-lock", fastIP)
	}
}
