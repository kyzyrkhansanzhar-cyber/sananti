package core

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileLogger_AsyncWritingAndGracefulShutdown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sananti-test-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logFilePath := filepath.Join(tmpDir, "alerts_test.log")

	// 1. Initialize FileLogger with a small buffer and high size boundary
	fl, err := NewFileLogger(logFilePath, 10, 10*1024*1024)
	if err != nil {
		t.Fatalf("failed to create FileLogger: %v", err)
	}

	alert := AlertData{
		IP:        "192.0.2.1",
		Path:      "/phpmyadmin",
		Method:    "POST",
		UserAgent: "Mozilla-Test",
		Severity:  SeverityCritical,
		Timestamp: time.Now().UTC(),
		Details:   "Intruder Decoy Triggered",
	}

	// 2. Log Alert
	if err := fl.LogAlert(alert); err != nil {
		t.Errorf("failed to log alert: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := fl.Close(ctx); err != nil {
		t.Fatalf("failed to gracefully close FileLogger: %v", err)
	}

	// 3. Verify file contents
	file, err := os.Open(logFilePath)
	if err != nil {
		t.Fatalf("failed to open written log file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var loggedAlert AlertData
	if err := decoder.Decode(&loggedAlert); err != nil {
		t.Fatalf("failed to decode alert from file: %v", err)
	}
	if loggedAlert.IP != alert.IP || loggedAlert.Path != alert.Path {
		t.Errorf("alert mismatch. Expected IP %s, got IP %s", alert.IP, loggedAlert.IP)
	}
}

func TestFileLogger_SizeBasedRotation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sananti-test-rotation-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logFilePath := filepath.Join(tmpDir, "rotation.log")

	// 1. Initialize FileLogger with a tiny size boundary (e.g. 50 bytes) to force instant rotation
	fl, err := NewFileLogger(logFilePath, 5, 50)
	if err != nil {
		t.Fatalf("failed to create FileLogger: %v", err)
	}

	alert1 := AlertData{
		IP:        "1.1.1.1",
		Path:      "/test",
		Method:    "GET",
		Severity:  SeverityInfo,
		Timestamp: time.Now().UTC(),
		Details:   "Trigger 1",
	}

	alert2 := AlertData{
		IP:        "2.2.2.2",
		Path:      "/test-2",
		Method:    "GET",
		Severity:  SeverityCritical,
		Timestamp: time.Now().UTC(),
		Details:   "Trigger 2",
	}

	// 2. Log first alert. It will write successfully and push the size over 50 bytes.
	if err := fl.LogAlert(alert1); err != nil {
		t.Fatalf("failed to log alert 1: %v", err)
	}

	// Sleep briefly to ensure the background worker has processed alert1
	time.Sleep(100 * time.Millisecond)

	// 3. Log second alert. It checks size, sees it exceeds 50 bytes, closes file, renames to rotation.log.1,
	// and writes the second alert to a new clean rotation.log.
	if err := fl.LogAlert(alert2); err != nil {
		t.Fatalf("failed to log alert 2: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := fl.Close(ctx); err != nil {
		t.Fatalf("failed to gracefully close FileLogger: %v", err)
	}

	// 4. Verify that rotation.log.1 exists (contains alert1)
	rotatedPath := logFilePath + ".1"
	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		t.Fatal("expected rotated backup file rotation.log.1 to exist, but it does not")
	}

	// Verify that the active rotation.log exists and contains alert2
	activeFile, err := os.Open(logFilePath)
	if err != nil {
		t.Fatalf("failed to open active log file: %v", err)
	}
	defer activeFile.Close()

	decoder := json.NewDecoder(activeFile)
	var loggedAlert2 AlertData
	if err := decoder.Decode(&loggedAlert2); err != nil {
		t.Fatalf("failed to decode active alert: %v", err)
	}
	if loggedAlert2.IP != "2.2.2.2" {
		t.Errorf("expected active log to contain alert 2 (IP: 2.2.2.2), got: %s", loggedAlert2.IP)
	}

	// Check EOF to ensure alert1 is NOT in the active log
	var dummy AlertData
	if err := decoder.Decode(&dummy); err != io.EOF {
		t.Errorf("expected active log to only contain 1 entry, but got more: %v", err)
	}
}
