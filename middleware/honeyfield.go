package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sananti/core"
)

// HoneyField coordinates form decoy traps (honeypot fields) and form time-lock
// validation to identify and block rapid automated spambots.
type HoneyField struct {
	blocker       core.Blocker
	logger        core.Logger
	fieldName     string
	secretKey     []byte // HMAC signing key for secure Time-Lock Tokens
	minSubmission time.Duration // Minimum allowed form-fill time (e.g. 800ms)
}

// NewHoneyField initializes a HoneyField configuration with form decoy fields and time-lock settings.
func NewHoneyField(blocker core.Blocker, logger core.Logger, fieldName string, minSubmission time.Duration) *HoneyField {
	return &HoneyField{
		blocker:       blocker,
		logger:        logger,
		fieldName:     fieldName,
		secretKey:     []byte("sananti_secure_timelock_key_2026"), // Standard default secret key
		minSubmission: minSubmission,
	}
}

// GenerateTimeLockToken generates a signed cryptographic token representing the current
// timestamp. This token is embedded as a hidden form field to prevent submission tampering.
func (hf *HoneyField) GenerateTimeLockToken() string {
	timestampStr := strconv.FormatInt(time.Now().UnixMilli(), 10)
	mac := hmac.New(sha256.New, hf.secretKey)
	mac.Write([]byte(timestampStr))
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", timestampStr, signature)
}

// verifyTimeLockToken validates the HMAC signature of the time-lock token and
// returns the elapsed duration since it was generated.
func (hf *HoneyField) verifyTimeLockToken(token string) (time.Duration, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid token format")
	}

	timestampStr, signature := parts[0], parts[1]

	// 1. Verify HMAC signature to prevent clients from forging timestamps
	mac := hmac.New(sha256.New, hf.secretKey)
	mac.Write([]byte(timestampStr))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return 0, fmt.Errorf("invalid signature signature")
	}

	// 2. Parse timestamp and calculate elapsed time
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp")
	}

	elapsed := time.Since(time.UnixMilli(timestamp))
	return elapsed, nil
}

// HandleField returns a middleware that checks POST, PUT, and PATCH request form values.
// It intercepts requests if either:
// 1. The invisible decoy honeypot input is filled out (Bot honeypot trigger).
// 2. The form is submitted faster than the minimum allowed time-lock threshold (Fast-Bot trigger).
func (hf *HoneyField) HandleField() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Decoy form checks apply only to state-modifying submissions
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				_ = r.ParseForm()
				_ = r.ParseMultipartForm(32 << 20)

				ip := ExtractIP(r)

				// --- Check 1: Decoy Honeypot Field Check ---
				if val := r.FormValue(hf.fieldName); val != "" {
					reason := fmt.Sprintf("Bot intrusion: decoy honeypot field %q filled with %q", hf.fieldName, val)
					
					alert := core.AlertData{
						IP:        ip,
						Path:      r.URL.Path,
						Method:    r.Method,
						UserAgent: r.UserAgent(),
						Severity:  core.SeverityCritical,
						Timestamp: time.Now(),
						Details:   reason,
					}
					_ = hf.logger.LogAlert(alert)
					_ = hf.blocker.BlockIP(ip, reason)

					http.Error(w, "403 Forbidden - Security Block", http.StatusForbidden)
					return
				}

				// --- Check 2: Form Submission Time-Lock Check ---
				token := r.FormValue("_sananti_timelock")
				if token == "" {
					// Missing token in a POST form suggests it wasn't requested via standard GET
					reason := "Suspicious submission: missing Time-Lock security token"
					alert := core.AlertData{
						IP:        ip,
						Path:      r.URL.Path,
						Method:    r.Method,
						UserAgent: r.UserAgent(),
						Severity:  core.SeverityWarning,
						Timestamp: time.Now(),
						Details:   reason,
					}
					_ = hf.logger.LogAlert(alert)
					// Block with warning
					_ = hf.blocker.BlockIP(ip, reason)

					http.Error(w, "403 Forbidden - Security Block", http.StatusForbidden)
					return
				}

				elapsed, err := hf.verifyTimeLockToken(token)
				if err != nil {
					reason := fmt.Sprintf("Suspicious submission: corrupted Time-Lock token: %v", err)
					alert := core.AlertData{
						IP:        ip,
						Path:      r.URL.Path,
						Method:    r.Method,
						UserAgent: r.UserAgent(),
						Severity:  core.SeverityWarning,
						Timestamp: time.Now(),
						Details:   reason,
					}
					_ = hf.logger.LogAlert(alert)
					_ = hf.blocker.BlockIP(ip, reason)

					http.Error(w, "403 Forbidden - Security Block", http.StatusForbidden)
					return
				}

				// If submission speed is abnormally fast, flag as automated bot
				if elapsed < hf.minSubmission {
					reason := fmt.Sprintf("Bot intrusion: rapid submission caught by Time-Lock (elapsed: %v, threshold: %v)", elapsed, hf.minSubmission)
					alert := core.AlertData{
						IP:        ip,
						Path:      r.URL.Path,
						Method:    r.Method,
						UserAgent: r.UserAgent(),
						Severity:  core.SeverityCritical,
						Timestamp: time.Now(),
						Details:   reason,
					}
					_ = hf.logger.LogAlert(alert)
					_ = hf.blocker.BlockIP(ip, reason)

					http.Error(w, "403 Forbidden - Security Block", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
