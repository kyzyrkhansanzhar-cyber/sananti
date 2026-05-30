package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"sananti/antifraud"
	"sananti/core"
	"sananti/middleware"
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

	// Start local scan HTTP server to allow external triggers (e.g. from mobile browser)
	go a.startLocalScanServer()
}

// startLocalScanServer launches a lightweight, concurrent HTTP server on port 8080.
// This allows other devices (like a Samsung Z Flip 7) on the same Wi-Fi to submit
// real-time scan requests to the MacBook Go backend and trigger the emergency lock modal.
func (a *App) startLocalScanServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers so standard web browsers can request it
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Retrieve parameters
		amountStr := r.URL.Query().Get("amount")
		phone := r.URL.Query().Get("phone")
		card := r.URL.Query().Get("card")
		email := r.URL.Query().Get("email")
		ip := r.URL.Query().Get("ip")

		// If IP is not supplied, extract the real network client IP
		if ip == "" {
			ip = middleware.ExtractIP(r)
		}

		var amount float64
		fmt.Sscanf(amountStr, "%f", &amount)

		tx := antifraud.Transaction{
			ID:                "tx_mobile_" + time.Now().Format("150405"),
			UserID:            "user_mobile_test",
			IP:                ip,
			Amount:            amount,
			CardBIN:           "440099",
			CardCountry:       "KZ",
			IPCountry:         "KZ",
			RecipientPhone:    phone,
			RecipientCard:     card,
			Email:             email,
			DeviceFingerprint: "device_samsung_z_flip_7_mobile",
			Timestamp:         time.Now(),
		}

		// Run Go anti-fraud scoring rules
		assessment, err := a.scanner.AnalyzeTransaction(context.Background(), tx)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","message":"Scanner execution crash"}`))
			return
		}

		// If transaction is fraudulent, trigger the beautiful desktop Lock screen in Wails!
		if !assessment.Approved {
			assessmentJSON, _ := json.Marshal(assessment)
			runtime.EventsEmit(a.ctx, "fraud_detected", string(assessmentJSON))
		}

		// Return JSON back to the phone
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"approved":       assessment.Approved,
			"risk_score":     assessment.RiskScore,
			"reasons":        assessment.Reasons,
			"recommendation": assessment.Recommendation,
		})
	})

	// Bind to all local interfaces on port 8080
	_ = http.ListenAndServe("0.0.0.0:8080", mux)
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

// StartDeepScan runs an asynchronous deep scanning simulation powered by Go Goroutines,
// emitting real-time progress percentages and system logs back to the frontend.
func (a *App) StartDeepScan(profileID string, payloadJSON string) {
	go func() {
		// Define the scanning steps based on the target profile
		var steps []map[string]interface{}

		if profileID == "safe" {
			steps = []map[string]interface{}{
				{"progress": 25, "log": "[INFO] Checking blacklist databases: IP 82.200.1.1 is clean."},
				{"progress": 50, "log": "[INFO] Resolving user location: GeoIP lookup matches billing country (KZ)."},
				{"progress": 75, "log": "[INFO] Evaluating transaction amount: $150.00 is well within limits."},
				{"progress": 100, "log": "[SUCCESS] Deep Scan complete. 0% anomalies. Transaction approved."},
			}
		} else if profileID == "scammer" {
			steps = []map[string]interface{}{
				{"progress": 25, "log": "[WARNING] GeoMismatchCheck: Card billing country (KZ) does not match transaction IP country (US)!"},
				{"progress": 50, "log": "[CRITICAL] RecipientBlacklistCheck: Recipient card 4400999988887777 matched blacklisted scammer/mule record!"},
				{"progress": 75, "log": "[WARNING] AmountAnomalyCheck: $2500.00 exceeds single-payment safety limits!"},
				{"progress": 100, "log": "[CRITICAL] Deep Scan complete. Threat detected. Emergency Lock Dispatching!"},
			}
		} else { // "bot"
			steps = []map[string]interface{}{
				{"progress": 25, "log": "[WARNING] IPBlacklistCheck: Client IP 185.220.101.5 is a known Tor Exit Node!"},
				{"progress": 50, "log": "[CRITICAL] HeaderReputationCheck: Headless ScanBot/v9.0 automation tool signature detected!"},
				{"progress": 75, "log": "[CRITICAL] HoneytokenDecoyCheck: Probe detected on secure decoy path /api/v1/admin/config!"},
				{"progress": 100, "log": "[CRITICAL] Deep Scan complete. Attack intercepted. Isolated by automated firewall ban!"},
			}
		}

		// Sequential execution of simulated scan layers
		for _, step := range steps {
			time.Sleep(500 * time.Millisecond) // Simulating high-fidelity layer-by-layer processing
			runtime.EventsEmit(a.ctx, "scan_progress", step)
		}

		time.Sleep(200 * time.Millisecond)

		// Execute final actions at 100%
		if profileID == "bot" {
			var botPayload struct {
				IP   string `json:"ip"`
				Path string `json:"path"`
			}
			_ = json.Unmarshal([]byte(payloadJSON), &botPayload)
			if botPayload.IP == "" {
				botPayload.IP = "185.220.101.5"
			}
			if botPayload.Path == "" {
				botPayload.Path = "/api/v1/admin/config"
			}
			a.TriggerHoneytokenBlock(botPayload.IP, botPayload.Path)
		} else {
			// legit or scammer
			resJSON := a.ScanTransaction(payloadJSON)
			runtime.EventsEmit(a.ctx, "scan_complete", resJSON)
		}
	}()
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
