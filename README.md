# Sananti: Go Deception-Defense & Anti-Fraud Security Library

Sananti is a lightweight, high-performance Go module designed to protect web applications against automated bots, vulnerability scanners, and payment fraud.

---

## Core Features

### 1. Active Threat Blocker (core)
* **Memory & Redis Support**: Thread-safe IP blocking with configurable expiration (TTL). Falls back to memory cache if Redis is offline.
* **CIDR Whitelisting**: Exclude corporate networks and safe subnets (e.g., `192.168.100.0/24`) from being blocked.

### 2. Honeytrap Decoy Interceptors (middleware)
* Registers deceptive paths (e.g., `/phpmyadmin`, `/api/v1/admin/config`) to trap reconnaissance bots.
* Instantly blacklists the offending IP address upon access.

### 3. Form Honeypot & Time-Lock Protection (middleware)
* **Decoy Fields**: Invisible form fields to intercept automated spam bots.
* **Cryptographic Time-Locks**: HMAC-signed tokens to block form submissions that occur too quickly (e.g., in less than 1 second).

### 4. Anti-Fraud Payment Scanner (antifraud)
* Evaluates transaction risk using multiple rules:
  - IP Blacklist & Country mismatch detection (IP vs Card Billing Country).
  - Velocity checks (frequency of transactions per user/device).
  - Disposable email domain detection.
  - Card BIN blacklist validation.
  - Device reputation tracking.

### 5. Management Dashboard & Telemetry
* Interactive web control panel running on port `:8081` to adjust thresholds and toggle scanning rules in real-time.
* Promotheus-compatible `/metrics` endpoint to monitor security events.

---

## Quick Start Integration

```go
package main

import (
	"log"
	"net/http"
	"time"
	"sananti/core"
	"sananti/middleware"
)

func main() {
	// 1. Initialize blocker with 1-hour default TTL
	blocker := core.NewMemoryBlocker(1 * time.Hour)
	_ = blocker.GetWhitelist().Add("10.0.0.0/8")

	// 2. Initialize rotation logger
	logger, _ := core.NewFileLogger("logs/security.log", 100, 5*1024*1024)

	// 3. Create honeytrap middleware
	honeyTrap := middleware.NewHoneyTrap(blocker, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Protected content"))
	})

	// Wrap application router
	var handler http.Handler = mux
	handler = honeyTrap.HandleTrap("/phpmyadmin")(handler)
	handler = honeyTrap.ProtectionMiddleware()(handler)

	log.Println("Server listening on :8081")
	http.ListenAndServe(":8081", handler)
}
```

---

## Тестілеу және іске қосу (KZ)
Жобадағы барлық тесттерді орындау:
```bash
go test -v ./...
```

Басқару панелі бар тестілік серверді іске қосу:
```bash
go run main.go
```
Браузерден ашыңыз: `http://localhost:8081/`

---

## Тестирование и Запуск (RU)
Запуск всех тестов:
```bash
go test -v ./...
```

Запуск тестового сервера управления:
```bash
go run main.go
```
Открыть в браузере: `http://localhost:8081/`
