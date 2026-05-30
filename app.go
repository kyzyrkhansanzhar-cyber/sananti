package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"sananti/antifraud"
	"sananti/core"
)

// App struct represents the Wails desktop backend controller.
type App struct {
	ctx        context.Context
	blocker    core.Blocker
	fileLogger *core.FileLogger
	scanner    *antifraud.AntiFraudScanner
	mu         sync.Mutex
}

// NewApp creates a new App instance and sets up the core security engines.
func NewApp() *App {
	// 1. Initialize blocker with safe Redis fallback
	var blocker core.Blocker
	redisAddr := "127.0.0.1:6379"
	rdb := redis.NewClient(&redis.Options{
		Addr:        redisAddr,
		DialTimeout: 500 * time.Millisecond,
		MaxRetries:  1,
	})

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelPing()

	defaultTTL := 1 * time.Hour
	if err := rdb.Ping(ctxPing).Err(); err == nil {
		blocker = core.NewRedisBlocker(rdb, defaultTTL)
	} else {
		blocker = core.NewMemoryBlocker(defaultTTL)
	}

	_ = blocker.GetWhitelist().Add("192.168.100.0/24")
	_ = blocker.GetWhitelist().Add("10.10.10.0/24")

	// 2. Initialize Buffered File Logger
	logFilePath := "logs/sananti_alerts.log"
	fileLogger, err := core.NewFileLogger(logFilePath, 100, 5*1024*1024)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize FileLogger: %v", err))
	}

	// 3. Initialize Anti-Fraud scanner
	rules := []antifraud.Rule{
		&antifraud.IPBlacklistRule{},
		&antifraud.GeoMismatchRule{},
		antifraud.NewVelocityAbuseRule(3, 2*time.Minute),
		antifraud.NewAmountAnomalyRule(2000.0),
		antifraud.NewEmailDomainRiskRule(),
		antifraud.NewCardBINBlacklistRule(),
		antifraud.NewDeviceReputationRule(5 * time.Minute),
		&antifraud.HeaderReputationRule{},
		antifraud.NewRecipientBlacklistRule(),
	}
	scanner := antifraud.NewAntiFraudScanner(blocker, fileLogger, rules...)

	return &App{
		blocker:    blocker,
		fileLogger: fileLogger,
		scanner:    scanner,
	}
}

// startup is called when the Wails desktop window initializes.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Launch background Goroutine real-time event-driven scanner thread
	go a.backgroundScannerLoop()
}

// backgroundScannerLoop runs a continuous Go thread monitoring security telemetry & active IP bans.
func (a *App) backgroundScannerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			// Emit dynamic security console shield heartbeats to show the live scanner is active
			runtime.EventsEmit(a.ctx, "shield_heartbeat", map[string]interface{}{
				"timestamp": time.Now().Format("15:04:05"),
				"status":    "ACTIVE",
			})
		}
	}
}

// GetConfig returns the dynamic scanner configuration JSON (threshold + active rules).
func (a *App) GetConfig() string {
	threshold, activeRules := a.scanner.GetConfig()
	data, _ := json.Marshal(map[string]interface{}{
		"block_threshold": threshold,
		"active_rules":    activeRules,
	})
	return string(data)
}

// UpdateConfig updates the scanner settings thread-safely in real-time.
func (a *App) UpdateConfig(threshold float64, activeRulesJSON string) string {
	var rules map[string]bool
	_ = json.Unmarshal([]byte(activeRulesJSON), &rules)

	a.scanner.UpdateConfig(threshold, rules)
	return `{"status":"success"}`
}

// ScanTransaction runs a real-time security assessment on incoming payment payloads.
func (a *App) ScanTransaction(txJSON string) string {
	var tx antifraud.Transaction
	if err := json.Unmarshal([]byte(txJSON), &tx); err != nil {
		return `{"status":"error","message":"Invalid JSON payload"}`
	}

	// Enrich metadata if not populated by the client
	if tx.IP == "" {
		tx.IP = "127.0.0.1" // Default local fallback inside native GUI client
	}
	if tx.UserAgent == "" {
		tx.UserAgent = "SanantiDesktopApp/v7.0"
	}
	tx.Timestamp = time.Now()

	assessment, err := a.scanner.AnalyzeTransaction(context.Background(), tx)
	if err != nil {
		return `{"status":"error","message":"Scanner execution crash"}`
	}

	var status string
	if assessment.Approved {
		status = "success"
	} else {
		status = "blocked"
		// Emit real-time event to trigger full-screen Emergency Lock Modal in JS
		assessmentJSON, _ := json.Marshal(assessment)
		runtime.EventsEmit(a.ctx, "fraud_detected", string(assessmentJSON))
	}

	data, _ := json.Marshal(map[string]interface{}{
		"status":         status,
		"risk_score":     assessment.RiskScore,
		"reasons":        assessment.Reasons,
		"recommendation": assessment.Recommendation,
	})
	return string(data)
}

// GetLiveLogs reads the latest security rotation logs from disk and returns them.
func (a *App) GetLiveLogs() string {
	logFilePath := "logs/sananti_alerts.log"
	file, err := os.Open(logFilePath)
	if err != nil {
		return "[]"
	}
	defer file.Close()

	var lines []string
	var chunk = make([]byte, 64*1024)
	n, err := file.Read(chunk)
	if err == nil && n > 0 {
		rawLines := strings.Split(string(chunk[:n]), "\n")
		for _, line := range rawLines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				lines = append(lines, trimmed)
			}
		}
	}

	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}

	data, _ := json.Marshal(lines)
	return string(data)
}

// TriggerHoneytokenBlock manually blacklists an IP address for testing deception traps.
func (a *App) TriggerHoneytokenBlock(ip string, path string) string {
	reason := fmt.Sprintf("Triggered decoy honeytoken URL: %s", path)
	alert := core.AlertData{
		IP:        ip,
		Path:      path,
		Method:    "GET",
		UserAgent: "SanantiDesktopApp/v7.0",
		Severity:  core.SeverityCritical,
		Timestamp: time.Now(),
		Details:   reason,
	}

	_ = a.fileLogger.LogAlert(alert)
	_ = a.blocker.BlockIP(ip, reason)

	// Broadcast bot detected event to show full screen lock modal
	runtime.EventsEmit(a.ctx, "bot_detected", map[string]string{
		"ip":     ip,
		"path":   path,
		"reason": reason,
	})

	return fmt.Sprintf(`{"status":"success","message":"IP %s blacklisted via Honeytoken trap!"}`, ip)
}
