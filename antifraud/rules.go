package antifraud

import (
	"fmt"
	"strings"
	"time"

	"sananti/core"
)

// IPBlacklistRule checks if the payment request's IP is already blacklisted in the core blocker.
type IPBlacklistRule struct{}

func (r *IPBlacklistRule) Name() string { return "IPBlacklistCheck" }

func (r *IPBlacklistRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if blocker == nil {
		return 0.0, ""
	}
	blocked, reason, err := blocker.IsBlocked(tx.IP)
	if err == nil && blocked {
		return 1.0, fmt.Sprintf("Client IP is blacklisted in core engine. Prior reason: %s", reason)
	}
	return 0.0, ""
}

// GeoMismatchRule identifies card-not-present fraud by comparing billing and IP locations.
type GeoMismatchRule struct{}

func (r *GeoMismatchRule) Name() string { return "GeoMismatchCheck" }

func (r *GeoMismatchRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if tx.CardCountry == "" || tx.IPCountry == "" {
		return 0.0, ""
	}
	if tx.CardCountry != tx.IPCountry {
		return 0.40, fmt.Sprintf("Card billing country (%s) does not match transaction IP country (%s)", tx.CardCountry, tx.IPCountry)
	}
	return 0.0, ""
}

// VelocityAbuseRule stops card-testing attacks by limiting short-term request counts per user.
type VelocityAbuseRule struct {
	MaxAttempts int
	Window      time.Duration
}

func NewVelocityAbuseRule(maxAttempts int, window time.Duration) *VelocityAbuseRule {
	return &VelocityAbuseRule{
		MaxAttempts: maxAttempts,
		Window:      window,
	}
}

func (r *VelocityAbuseRule) Name() string { return "VelocityAbuseCheck" }

func (r *VelocityAbuseRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	cutoff := tx.Timestamp.Add(-r.Window)
	recentAttempts := 0

	for _, prevTx := range history {
		if prevTx.Timestamp.After(cutoff) && prevTx.UserID == tx.UserID {
			recentAttempts++
		}
	}

	if recentAttempts >= r.MaxAttempts {
		return 0.75, fmt.Sprintf("High transaction frequency: %d payments attempted within the last %s", recentAttempts, r.Window)
	}
	return 0.0, ""
}

// AmountAnomalyRule flags unusually large transactions that exceed high-value limits.
type AmountAnomalyRule struct {
	Limit float64
}

func NewAmountAnomalyRule(limit float64) *AmountAnomalyRule {
	return &AmountAnomalyRule{
		Limit: limit,
	}
}

func (r *AmountAnomalyRule) Name() string { return "AmountAnomalyCheck" }

func (r *AmountAnomalyRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if tx.Amount >= r.Limit {
		return 0.40, fmt.Sprintf("Transaction amount (%.2f) exceeds single-payment high-value safety limit (%.2f)", tx.Amount, r.Limit)
	}
	return 0.0, ""
}

// EmailDomainRiskRule flags transactions submitted with temporary or disposable email addresses.
type EmailDomainRiskRule struct {
	disposableDomains map[string]bool
}

func NewEmailDomainRiskRule() *EmailDomainRiskRule {
	domains := map[string]bool{
		"mailinator.com":      true,
		"yopmail.com":         true,
		"tempmail.com":        true,
		"10minutemail.com":    true,
		"guerrillamail.com":   true,
		"dispostable.com":     true,
		"sharklasers.com":     true,
		"getairmail.com":      true,
	}
	return &EmailDomainRiskRule{disposableDomains: domains}
}

func (r *EmailDomainRiskRule) Name() string { return "EmailDomainRiskCheck" }

func (r *EmailDomainRiskRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if tx.Email == "" {
		return 0.0, ""
	}

	parts := strings.Split(tx.Email, "@")
	if len(parts) != 2 {
		return 0.0, ""
	}

	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	if r.disposableDomains[domain] {
		return 0.35, fmt.Sprintf("Disposable/Temporary email domain detected: %q", domain)
	}

	return 0.0, ""
}

// CardBINBlacklistRule flags credit cards associated with high-risk prepaid ranges or known test BINs.
type CardBINBlacklistRule struct {
	blockedBINs map[string]bool
}

func NewCardBINBlacklistRule() *CardBINBlacklistRule {
	// Example BINs representing high-risk or card-testing ranges
	bins := map[string]bool{
		"400011": true, // Classic test visa BIN
		"411111": true, // Classic visa test card
		"422222": true, // Prepaid block BIN
		"555555": true, // Mastercard test card
	}
	return &CardBINBlacklistRule{blockedBINs: bins}
}

func (r *CardBINBlacklistRule) Name() string { return "CardBINBlacklistCheck" }

func (r *CardBINBlacklistRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if len(tx.CardBIN) < 6 {
		return 0.0, ""
	}

	bin := tx.CardBIN[:6]
	if r.blockedBINs[bin] {
		return 0.40, fmt.Sprintf("High-risk or card-testing Credit Card BIN detected: %q", bin)
	}

	return 0.0, ""
}

// DeviceReputationRule tracks fingerprint velocity. If the same device fingerprint
// attempts payments with different emails or users in short periods, it flags account-takeover.
type DeviceReputationRule struct {
	Window time.Duration
}

func NewDeviceReputationRule(window time.Duration) *DeviceReputationRule {
	return &DeviceReputationRule{Window: window}
}

func (r *DeviceReputationRule) Name() string { return "DeviceReputationCheck" }

func (r *DeviceReputationRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if tx.DeviceFingerprint == "" {
		return 0.0, ""
	}

	cutoff := tx.Timestamp.Add(-r.Window)
	emailsUsed := make(map[string]bool)
	emailsUsed[tx.Email] = true

	for _, prevTx := range history {
		if prevTx.Timestamp.After(cutoff) && prevTx.DeviceFingerprint == tx.DeviceFingerprint {
			if prevTx.Email != "" {
				emailsUsed[prevTx.Email] = true
			}
		}
	}

	// Trigger risk if the same device is used with 3 or more distinct email addresses
	if len(emailsUsed) >= 3 {
		return 0.60, fmt.Sprintf("Suspicious device fingerprint: %d distinct emails used on this device in under %s", len(emailsUsed), r.Window)
	}

	return 0.0, ""
}

// HeaderReputationRule detects automated headless browsers, curl scripts, and testing tools.
type HeaderReputationRule struct{}

func (r *HeaderReputationRule) Name() string { return "HeaderReputationCheck" }

func (r *HeaderReputationRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if tx.UserAgent == "" {
		return 0.0, ""
	}

	ua := strings.ToLower(tx.UserAgent)
	suspiciousSignatures := []string{
		"headlesschrome",
		"puppeteer",
		"selenium",
		"playwright",
		"python-requests",
		"curl/",
		"wget/",
		"postmanruntime",
		"http-client",
	}

	for _, sig := range suspiciousSignatures {
		if strings.Contains(ua, sig) {
			return 0.50, fmt.Sprintf("Automated user-agent or headless script signature detected: %q", sig)
		}
	}

	return 0.0, ""
}

// RecipientBlacklistRule blocks transactions sent to known scammers, fraudsters, or mule accounts.
type RecipientBlacklistRule struct {
	scamPhones map[string]bool
	scamCards  map[string]bool
}

func NewRecipientBlacklistRule() *RecipientBlacklistRule {
	phones := map[string]bool{
		"+77777777777": true,
		"87777777777":  true,
		"+77079998877": true,
		"87079998877":  true,
		"+77779998811": true,
		"87779998811":  true,
	}
	cards := map[string]bool{
		"4400551122334455": true,
		"4111112222222222": true,
		"4400999988887777": true,
	}
	return &RecipientBlacklistRule{
		scamPhones: phones,
		scamCards:  cards,
	}
}

func (r *RecipientBlacklistRule) Name() string { return "RecipientBlacklistCheck" }

func (r *RecipientBlacklistRule) Evaluate(tx Transaction, history []Transaction, blocker core.Blocker) (float64, string) {
	if tx.RecipientPhone != "" {
		cleanedPhone := strings.ReplaceAll(tx.RecipientPhone, " ", "")
		if r.scamPhones[cleanedPhone] {
			return 1.0, fmt.Sprintf("Recipient phone number %s is blacklisted as a known SCAMMER", cleanedPhone)
		}
	}

	if tx.RecipientCard != "" {
		cleanedCard := strings.ReplaceAll(tx.RecipientCard, " ", "")
		if r.scamCards[cleanedCard] {
			return 1.0, fmt.Sprintf("Recipient card number %s is blacklisted as a known SCAM/MULE account", cleanedCard)
		}
	}

	return 0.0, ""
}


