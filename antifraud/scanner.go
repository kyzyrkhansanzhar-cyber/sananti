package antifraud

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sananti/core"
)

// Transaction represents a financial transaction captured at the exact moment of a payment.
type Transaction struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	IP                string    `json:"ip"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	CardBIN           string    `json:"card_bin"`
	CardCountry       string    `json:"card_country"`
	IPCountry         string    `json:"ip_country"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	Email             string    `json:"email"`
	Timestamp         time.Time `json:"timestamp"`
}

// RiskAssessment describes the security output of the anti-fraud scan.
type RiskAssessment struct {
	Approved       bool      `json:"approved"`       // False if the transaction is blocked
	RiskScore      float64   `json:"risk_score"`     // 0.0 (Safe) to 1.0 (Highly Fraudulent)
	Reasons        []string  `json:"reasons"`        // Specific rule violations triggered
	Recommendation string    `json:"recommendation"` // "APPROVE", "REVIEW", "BLOCK"
}

// Rule defines the contract for implementing customized transaction evaluation rules.
type Rule interface {
	Name() string
	Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (scoreContribution float64, reason string)
}

// AntiFraudScanner operates as the proactive digital guard. It records transaction history
// and runs rules engines in real-time to intercept and block fraudulent money transfers.
type AntiFraudScanner struct {
	mu             sync.RWMutex
	blocker        core.Blocker
	logger         core.Logger
	rules          []Rule
	txHistory      map[string][]Transaction // Maps UserID to recent transactions
	blockThreshold float64                  // Dynamic threshold for hard blocks (default: 0.70)
	activeRules    map[string]bool          // Dynamic toggles for specific rules (default: all active)
}

// NewAntiFraudScanner instantiates a new AntiFraudScanner with custom rules and default config.
func NewAntiFraudScanner(blocker core.Blocker, logger core.Logger, rules ...Rule) *AntiFraudScanner {
	activeMap := make(map[string]bool)
	for _, r := range rules {
		activeMap[r.Name()] = true
	}

	return &AntiFraudScanner{
		blocker:        blocker,
		logger:         logger,
		rules:          rules,
		txHistory:      make(map[string][]Transaction),
		blockThreshold: 0.70,
		activeRules:    activeMap,
	}
}

// GetConfig returns the thread-safe copy of current scanner configuration boundaries.
func (afs *AntiFraudScanner) GetConfig() (float64, map[string]bool) {
	afs.mu.RLock()
	defer afs.mu.RUnlock()

	// Perform map copy to prevent external concurrency races
	rulesCopy := make(map[string]bool, len(afs.activeRules))
	for k, v := range afs.activeRules {
		rulesCopy[k] = v
	}

	return afs.blockThreshold, rulesCopy
}

// UpdateConfig updates the blocking threshold and active rules map in real-time.
func (afs *AntiFraudScanner) UpdateConfig(threshold float64, active map[string]bool) {
	afs.mu.Lock()
	defer afs.mu.Unlock()

	afs.blockThreshold = threshold
	if active != nil {
		afs.activeRules = active
	}
}

// AnalyzeTransaction executes all active fraud rules in parallel or sequential order.
// If the risk score exceeds the dynamic blockThreshold, it blocks the transaction.
func (afs *AntiFraudScanner) AnalyzeTransaction(ctx context.Context, tx Transaction) (RiskAssessment, error) {
	afs.mu.Lock()
	if tx.Timestamp.IsZero() {
		tx.Timestamp = time.Now()
	}
	history := afs.txHistory[tx.UserID]
	afs.txHistory[tx.UserID] = append(history, tx)

	// Fetch config parameters thread-safely
	threshold := afs.blockThreshold
	activeMap := make(map[string]bool, len(afs.activeRules))
	for k, v := range afs.activeRules {
		activeMap[k] = v
	}
	afs.mu.Unlock()

	var cumulativeScore float64
	var triggeredReasons []string

	// Evaluate all active rules
	for _, rule := range afs.rules {
		// Only run the rule if it is dynamically toggled ON
		if active, exists := activeMap[rule.Name()]; exists && !active {
			continue
		}

		score, reason := rule.Evaluate(tx, history, afs.blocker)
		if score > 0 {
			cumulativeScore += score
			triggeredReasons = append(triggeredReasons, fmt.Sprintf("[%s]: %s (Risk: +%.2f)", rule.Name(), reason, score))

			// Update core Prometheus telemetry collectors based on matching rules
			switch rule.Name() {
			case "GeoMismatchCheck":
				core.GlobalTelemetry.IncrementGeoMismatches()
			case "VelocityAbuseCheck":
				core.GlobalTelemetry.IncrementVelocityBlocks()
			case "AmountAnomalyCheck":
				core.GlobalTelemetry.IncrementAmountBlocks()
			case "EmailDomainRiskCheck":
				core.GlobalTelemetry.IncrementDisposableEmails()
			}
		}
	}

	if cumulativeScore > 1.0 {
		cumulativeScore = 1.0
	}

	assessment := RiskAssessment{
		RiskScore: cumulativeScore,
		Reasons:   triggeredReasons,
	}

	// Dynamic comparison against threshold
	switch {
	case cumulativeScore >= threshold:
		assessment.Approved = false
		assessment.Recommendation = "BLOCK"
		core.GlobalTelemetry.IncrementBlockedIPs()
	case cumulativeScore >= 0.35:
		assessment.Approved = true
		assessment.Recommendation = "REVIEW"
	default:
		assessment.Approved = true
		assessment.Recommendation = "APPROVE"
	}

	// Log a high-severity alert asynchronously if transaction is suspicious or blocked
	if !assessment.Approved || assessment.Recommendation == "REVIEW" {
		alertDetail := fmt.Sprintf("Anti-Fraud Flagged: %s. Reasons: %v", assessment.Recommendation, triggeredReasons)
		alert := core.AlertData{
			IP:        tx.IP,
			Path:      "/api/v1/payment",
			Method:    "POST",
			UserAgent: "AntiFraudScanner",
			Severity:  core.SeverityCritical,
			Timestamp: time.Now(),
			Details:   alertDetail,
		}
		if !assessment.Approved {
			alert.Severity = core.SeverityCritical
		} else {
			alert.Severity = core.SeverityWarning
		}

		_ = afs.logger.LogAlert(alert)

		if !assessment.Approved {
			_ = afs.blocker.BlockIP(tx.IP, fmt.Sprintf("Anti-Fraud Payment Blocked. Details: %s", alertDetail))
		}
	}

	return assessment, nil
}

// ClearHistory removes cached transaction histories older than the duration limit.
func (afs *AntiFraudScanner) ClearHistory(olderThan time.Duration) {
	afs.mu.Lock()
	defer afs.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for userID, txs := range afs.txHistory {
		var active []Transaction
		for _, tx := range txs {
			if tx.Timestamp.After(cutoff) {
				active = append(active, tx)
			}
		}
		if len(active) == 0 {
			delete(afs.txHistory, userID)
		} else {
			afs.txHistory[userID] = active
		}
	}
}
