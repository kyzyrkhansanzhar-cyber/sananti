package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"sananti/antifraud"
	"sananti/core"
)

// AntiFraudMiddleware provides automated zero-touch background scanning of incoming requests.
type AntiFraudMiddleware struct {
	scanner *antifraud.AntiFraudScanner
	blocker core.Blocker
	logger  core.Logger
}

// NewAntiFraudMiddleware instantiates the background antivirus middleware.
func NewAntiFraudMiddleware(scanner *antifraud.AntiFraudScanner, blocker core.Blocker, logger core.Logger) *AntiFraudMiddleware {
	return &AntiFraudMiddleware{
		scanner: scanner,
		blocker: blocker,
		logger:  logger,
	}
}

// AutoScanMiddleware intercepts POST requests, automatically extracts recipient data,
// evaluates fraud risk scores in the background, and dynamically blocks scammers before they can take money.
func (afm *AntiFraudMiddleware) AutoScanMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only scan POST requests which usually contain payment/transfer payloads
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Read request body non-destructively
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			// Restore request body so that subsequent handlers can still read it cleanly
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Decode temporary map to inspect for potential transaction keys
			var payload map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &payload); err != nil {
				// If not valid JSON, proceed normally (e.g. standard form submissions)
				next.ServeHTTP(w, r)
				return
			}

			// Check if payload contains transaction fields (either recipient_phone or recipient_card)
			_, hasPhone := payload["recipient_phone"].(string)
			_, hasCard := payload["recipient_card"].(string)

			if hasPhone || hasCard {
				// Construct a transaction object automatically in the background
				var tx antifraud.Transaction
				_ = json.Unmarshal(bodyBytes, &tx)

				// Enrich IP and User-Agent parameters securely if missing
				if tx.IP == "" {
					tx.IP = ExtractIP(r)
				}
				if tx.UserAgent == "" {
					tx.UserAgent = r.UserAgent()
				}
				if tx.Timestamp.IsZero() {
					tx.Timestamp = time.Now()
				}

				// Run background anti-fraud scan automatically
				assessment, err := afm.scanner.AnalyzeTransaction(r.Context(), tx)
				if err == nil && !assessment.Approved {
					// Dynamic block triggered: return 403 Forbidden with details immediately
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					resp, _ := json.Marshal(map[string]interface{}{
						"status":         "blocked",
						"error":          "FRAUD_DETECTED",
						"risk_score":     assessment.RiskScore,
						"reasons":        assessment.Reasons,
						"recommendation": assessment.Recommendation,
						"message":        "Төлем бұғатталды! Алушы алаяқтардың қара тізімінде бар.",
					})
					_, _ = w.Write(resp)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
