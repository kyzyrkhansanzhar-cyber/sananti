package core

import (
	"fmt"
	"sync"
)

// TelemetryCollector tracks global security metrics and exports them in
// Prometheus-compatible format for Grafana and dashboard auditing.
type TelemetryCollector struct {
	mu               sync.Mutex
	blockedIPs       int64
	honeypotHits     int64
	disposableEmails int64
	geoMismatches    int64
	velocityBlocks   int64
	amountBlocks     int64
}

var (
	// GlobalTelemetry is a singleton instance of our thread-safe metrics collector.
	GlobalTelemetry = NewTelemetryCollector()
)

// NewTelemetryCollector initializes a clean TelemetryCollector.
func NewTelemetryCollector() *TelemetryCollector {
	return &TelemetryCollector{}
}

// IncrementBlockedIPs increments the total count of blacklisted IPs.
func (t *TelemetryCollector) IncrementBlockedIPs() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.blockedIPs++
}

// IncrementHoneypotHits increments the number of decoy URL/field hits.
func (t *TelemetryCollector) IncrementHoneypotHits() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.honeypotHits++
}

// IncrementDisposableEmails increments the count of disposable email detections.
func (t *TelemetryCollector) IncrementDisposableEmails() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.disposableEmails++
}

// IncrementGeoMismatches increments billing/IP location mismatch events.
func (t *TelemetryCollector) IncrementGeoMismatches() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.geoMismatches++
}

// IncrementVelocityBlocks increments payment rate-limiting triggers.
func (t *TelemetryCollector) IncrementVelocityBlocks() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.velocityBlocks++
}

// IncrementAmountBlocks flags payments blocked due to large value anomalies.
func (t *TelemetryCollector) IncrementAmountBlocks() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.amountBlocks++
}

// PrometheusFormat outputs standard, scrape-ready Prometheus text metrics.
func (t *TelemetryCollector) PrometheusFormat() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	return fmt.Sprintf(
		"# HELP sananti_blocked_ips_total Total number of blacklisted IP addresses.\n"+
			"# TYPE sananti_blocked_ips_total counter\n"+
			"sananti_blocked_ips_total %d\n\n"+
			"# HELP sananti_honeypot_hits_total Total number of deceptive honeypot and decoy hits.\n"+
			"# TYPE sananti_honeypot_hits_total counter\n"+
			"sananti_honeypot_hits_total %d\n\n"+
			"# HELP sananti_disposable_emails_total Total number of disposable email domain detections.\n"+
			"# TYPE sananti_disposable_emails_total counter\n"+
			"sananti_disposable_emails_total %d\n\n"+
			"# HELP sananti_geo_mismatches_total Total card/IP location discrepancies flagged.\n"+
			"# TYPE sananti_geo_mismatches_total counter\n"+
			"sananti_geo_mismatches_total %d\n\n"+
			"# HELP sananti_velocity_blocks_total Total number of payments blocked by rate-limiting (velocity).\n"+
			"# TYPE sananti_velocity_blocks_total counter\n"+
			"sananti_velocity_blocks_total %d\n\n"+
			"# HELP sananti_amount_blocks_total Total payments blocked due to single-value limits.\n"+
			"# TYPE sananti_amount_blocks_total counter\n"+
			"sananti_amount_blocks_total %d\n",
		t.blockedIPs, t.honeypotHits, t.disposableEmails, t.geoMismatches, t.velocityBlocks, t.amountBlocks,
	)
}
