package core

import "time"

// AlertSeverity defines the risk classification levels for security log entries.
type AlertSeverity string

const (
	// SeverityInfo represents standard events or decoy scans with low immediate threat.
	SeverityInfo AlertSeverity = "INFO"

	// SeverityWarning represents suspicious activities like repeated scans or card review flags.
	SeverityWarning AlertSeverity = "WARNING"

	// SeverityCritical represents confirmed malicious blocks or direct transaction intercepts.
	SeverityCritical AlertSeverity = "CRITICAL"
)

// BlockInfo contains details about a blacklisted IP address, including its expiration.
type BlockInfo struct {
	BlockedAt time.Time `json:"blocked_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"` // Automatic unblocking time
	Reason    string    `json:"reason"`
	Attempts  int       `json:"attempts"`
}

// AlertData holds rich structured contextual information about detected intrusions.
type AlertData struct {
	IP        string        `json:"ip"`
	Path      string        `json:"path"`
	Method    string        `json:"method"`
	UserAgent string        `json:"user_agent"`
	Severity  AlertSeverity `json:"severity"`
	Timestamp time.Time     `json:"timestamp"`
	Details   string        `json:"details"`
}
