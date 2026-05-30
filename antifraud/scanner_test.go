package antifraud

import (
	"context"
	"testing"
	"time"

	"sananti/core"
)

type mockLogger struct{}

func (m *mockLogger) LogAlert(alert core.AlertData) error {
	return nil
}

func TestAntiFraudScanner_RulesEvaluation(t *testing.T) {
	blocker := core.NewMemoryBlocker(1 * time.Hour)
	defer blocker.Close()
	mlog := &mockLogger{}

	// Setup our scanner with standard rule thresholds
	rules := []Rule{
		&IPBlacklistRule{},
		&GeoMismatchRule{},
		NewVelocityAbuseRule(3, 10*time.Second),
		NewAmountAnomalyRule(1000.0),
		NewEmailDomainRiskRule(),
		NewCardBINBlacklistRule(),
		NewDeviceReputationRule(10 * time.Second),
		&HeaderReputationRule{},
		NewRecipientBlacklistRule(),
	}
	scanner := NewAntiFraudScanner(blocker, mlog, rules...)

	ctx := context.Background()

	// --- Case 1: Purely Safe Transaction ---
	txSafe := Transaction{
		ID:                "tx-1",
		UserID:            "user-1",
		IP:                "192.168.1.10",
		Amount:            100.0,
		CardCountry:       "KZ",
		IPCountry:         "KZ",
		CardBIN:           "440055", // Normal BIN
		Email:             "alice@gmail.com",
		DeviceFingerprint: "fingerprint_1",
		Timestamp:         time.Now(),
	}
	assess, err := scanner.AnalyzeTransaction(ctx, txSafe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !assess.Approved {
		t.Errorf("expected safe transaction to be approved")
	}
	if assess.RiskScore != 0.0 {
		t.Errorf("expected safe transaction risk score to be 0.0, got %.2f", assess.RiskScore)
	}
	if assess.Recommendation != "APPROVE" {
		t.Errorf("expected recommendation 'APPROVE', got %q", assess.Recommendation)
	}

	// --- Case 2: Disposable Email Risk ---
	txEmail := Transaction{
		ID:          "tx-2",
		UserID:      "user-2",
		IP:          "192.168.1.20",
		Amount:      50.0,
		CardCountry: "KZ",
		IPCountry:   "KZ",
		CardBIN:     "440055",
		Email:       "scammer@mailinator.com", // Disposable domain
		Timestamp:   time.Now(),
	}
	assessEmail, err := scanner.AnalyzeTransaction(ctx, txEmail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !assessEmail.Approved {
		t.Error("expected disposable email alone to not block instantly (score: 0.35)")
	}
	if assessEmail.RiskScore != 0.35 {
		t.Errorf("expected risk score 0.35, got %.2f", assessEmail.RiskScore)
	}
	if assessEmail.Recommendation != "REVIEW" {
		t.Errorf("expected recommendation 'REVIEW', got %q", assessEmail.Recommendation)
	}

	// --- Case 3: Blacklisted Card BIN Check ---
	txBIN := Transaction{
		ID:          "tx-3",
		UserID:      "user-3",
		IP:          "192.168.1.30",
		Amount:      80.0,
		CardCountry: "KZ",
		IPCountry:   "KZ",
		CardBIN:     "400011", // Blacklisted card testing BIN
		Email:       "bob@gmail.com",
		Timestamp:   time.Now(),
	}
	assessBIN, err := scanner.AnalyzeTransaction(ctx, txBIN)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !assessBIN.Approved {
		t.Error("expected blacklisted BIN alone to flag REVIEW (score: 0.40)")
	}
	if assessBIN.RiskScore != 0.40 {
		t.Errorf("expected risk score 0.40, got %.2f", assessBIN.RiskScore)
	}

	// --- Case 4: Device Reputation Velocity Check (Account Takeover) ---
	// Same device fingerprint, 3 different emails in under 10 seconds
	deviceFP := "fraud_device_device_99"
	
	// Tx A - Email 1
	_, _ = scanner.AnalyzeTransaction(ctx, Transaction{ID: "dev-1", UserID: "user-dev-abuse", IP: "1.1.1.1", Amount: 10.0, Email: "a@gmail.com", DeviceFingerprint: deviceFP, Timestamp: time.Now()})
	// Tx B - Email 2
	_, _ = scanner.AnalyzeTransaction(ctx, Transaction{ID: "dev-2", UserID: "user-dev-abuse", IP: "1.1.1.1", Amount: 10.0, Email: "b@gmail.com", DeviceFingerprint: deviceFP, Timestamp: time.Now()})
	
	// Tx C - Email 3 (Triggers DeviceReputation check score +0.60)
	txDeviceAbuse := Transaction{
		ID:                "dev-3",
		UserID:            "user-dev-abuse",
		IP:                "1.1.1.1",
		Amount:            10.0,
		Email:             "c@gmail.com",
		DeviceFingerprint: deviceFP,
		Timestamp:         time.Now(),
	}
	assessDevice, err := scanner.AnalyzeTransaction(ctx, txDeviceAbuse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !assessDevice.Approved {
		// Single device reputation score is 0.60, falls into REVIEW.
		// If combined with other rules (e.g. Card BIN or disposable email), it blocks!
		t.Logf("Device abuse approved but flagged. Score: %.2f Recommendation: %s", assessDevice.RiskScore, assessDevice.Recommendation)
	}
	if assessDevice.RiskScore != 0.60 {
		t.Errorf("expected risk score 0.60, got %.2f", assessDevice.RiskScore)
	}

	// --- Case 5: Automated User Agent Header Check ---
	txUA := Transaction{
		ID:        "tx-ua-1",
		UserID:    "user-ua",
		IP:        "192.168.1.50",
		Amount:    30.0,
		UserAgent: "Mozilla/5.0 HeadlessChrome/124.0.0.0 Safari/5.37", // Suspicious headless user-agent
		Timestamp: time.Now(),
	}
	assessUA, err := scanner.AnalyzeTransaction(ctx, txUA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !assessUA.Approved {
		t.Error("expected automated user-agent alone to flag REVIEW (score: 0.50)")
	}
	if assessUA.RiskScore != 0.50 {
		t.Errorf("expected risk score 0.50, got %.2f", assessUA.RiskScore)
	}

	// --- Case 6: Recipient Blacklist Scam Number Check ---
	txScam := Transaction{
		ID:             "tx-scam-1",
		UserID:         "user-victim",
		IP:             "192.168.1.10",
		Amount:         200.0,
		RecipientPhone: "+77777777777", // Blacklisted scam number!
		Timestamp:      time.Now(),
	}
	assessScam, err := scanner.AnalyzeTransaction(ctx, txScam)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assessScam.Approved || assessScam.Recommendation != "BLOCK" {
		t.Error("expected scam recipient to trigger instant BLOCK")
	}
	if assessScam.RiskScore != 1.0 {
		t.Errorf("expected risk score 1.0, got %.2f", assessScam.RiskScore)
	}
}

func TestAntiFraudScanner_DynamicConfig(t *testing.T) {
	blocker := core.NewMemoryBlocker(1 * time.Hour)
	defer blocker.Close()
	mlog := &mockLogger{}

	rules := []Rule{
		&GeoMismatchRule{},
	}
	scanner := NewAntiFraudScanner(blocker, mlog, rules...)
	ctx := context.Background()

	txGeo := Transaction{
		ID:          "tx-geo-1",
		UserID:      "user-config",
		IP:          "192.168.1.10",
		Amount:      100.0,
		CardCountry: "US",
		IPCountry:   "KZ",
		Timestamp:   time.Now(),
	}

	// 1. Initial State: Threshold is 0.70. Geo mismatch risk score is 0.40.
	// Since 0.40 < 0.70, it must be APPROVED but flagged for REVIEW.
	assessInit, _ := scanner.AnalyzeTransaction(ctx, txGeo)
	if !assessInit.Approved || assessInit.Recommendation != "REVIEW" {
		t.Errorf("expected transaction to be approved with REVIEW recommendation, got Approved=%t, Rec=%q",
			assessInit.Approved, assessInit.Recommendation)
	}

	// 2. Test Dynamic Threshold Update: Lower block threshold to 0.30
	threshold, rulesMap := scanner.GetConfig()
	if threshold != 0.70 {
		t.Errorf("expected default threshold to be 0.70, got %.2f", threshold)
	}

	scanner.UpdateConfig(0.30, rulesMap)

	// Since 0.40 >= 0.30, the transaction must now be BLOCKED instantly!
	assessLowered, _ := scanner.AnalyzeTransaction(ctx, txGeo)
	if assessLowered.Approved || assessLowered.Recommendation != "BLOCK" {
		t.Errorf("expected transaction to be blocked due to lowered threshold, got Approved=%t, Rec=%q",
			assessLowered.Approved, assessLowered.Recommendation)
	}

	// 3. Test Dynamic Rules Toggle: Disable the "GeoMismatchCheck" rule
	rulesMap["GeoMismatchCheck"] = false
	scanner.UpdateConfig(0.30, rulesMap)

	// Since the rule is disabled, the risk score must be 0.00 and thus APPROVED!
	assessDisabled, _ := scanner.AnalyzeTransaction(ctx, txGeo)
	if !assessDisabled.Approved || assessDisabled.RiskScore != 0.00 {
		t.Errorf("expected transaction to be approved with risk score 0.00 since rule is disabled, got Approved=%t, Score=%.2f",
			assessDisabled.Approved, assessDisabled.RiskScore)
	}
}

