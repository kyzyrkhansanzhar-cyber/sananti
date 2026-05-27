package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Logger defines the behavioral contract for recording intrusion alerts.
type Logger interface {
	LogAlert(alert AlertData) error
}

// FileLogger is an asynchronous channel-based file logger that processes
// alerts in a background worker goroutine and supports size-based log rotation.
type FileLogger struct {
	file     *os.File
	filePath string
	maxSize  int64 // Maximum log file size in bytes before triggering rotation
	ch       chan AlertData
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	closed   bool
	mu       sync.Mutex
}

// NewFileLogger creates a production-ready asynchronous logger with auto-rotation.
// It creates any missing parent directories safely across all operating systems.
func NewFileLogger(filePath string, bufferSize int, maxSize int64) (*FileLogger, error) {
	cleanPath := filepath.Clean(filePath)

	// Ensure log directory path exists
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create the log file in append mode
	file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	fl := &FileLogger{
		file:     file,
		filePath: cleanPath,
		maxSize:  maxSize,
		ch:       make(chan AlertData, bufferSize),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start the asynchronous log processor worker
	fl.wg.Add(1)
	go fl.worker()

	return fl, nil
}

// LogAlert pushes the alert metadata onto the channel buffer for asynchronous writing.
func (fl *FileLogger) LogAlert(alert AlertData) error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	if fl.closed {
		return fmt.Errorf("logger is closed")
	}

	// Non-blocking select to prevent blocking payment threads if buffer is full
	select {
	case fl.ch <- alert:
		return nil
	default:
		_, _ = fmt.Fprintf(os.Stderr, "[Sananti Alert Buffer Full] Dropped alert from IP: %s\n", alert.IP)
		return fmt.Errorf("alert channel buffer is full, alert dropped")
	}
}

// worker reads from the channel and writes to the file descriptor sequentially.
func (fl *FileLogger) worker() {
	defer fl.wg.Done()

	for alert := range fl.ch {
		// Perform log file size rotation check before writing
		if err := fl.checkRotation(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to check/rotate log file: %v\n", err)
		}

		data, err := json.Marshal(alert)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to marshal alert: %v\n", err)
			continue
		}

		_, err = fl.file.Write(append(data, '\n'))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to write alert to file: %v\n", err)
		}
	}
}

// checkRotation monitors the current log file size. If it exceeds the maxSize boundary,
// it rotates the log to a backup file atomically and reopens a fresh file handle.
func (fl *FileLogger) checkRotation() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// Get current file size information
	info, err := fl.file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	// Trigger rotation if file size exceeds our maximum configured boundary
	if info.Size() >= fl.maxSize {
		// 1. Close current active file descriptor
		_ = fl.file.Close()

		// 2. Perform rotation by renaming active log to backup path
		backupPath := fl.filePath + ".1"
		_ = os.Remove(backupPath) // Delete old backup if it exists
		if err := os.Rename(fl.filePath, backupPath); err != nil {
			// Reopen original file on failure to prevent logger hang
			fl.file, _ = os.OpenFile(fl.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			return fmt.Errorf("failed to rename log file during rotation: %w", err)
		}

		// 3. Open a fresh active log file descriptor
		newFile, err := os.OpenFile(fl.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open fresh log file: %w", err)
		}

		fl.file = newFile
	}

	return nil
}

// Close gracefully shuts down the logger. It closes the channel, allowing the worker
// to finish writing all queued alerts, then closes the file. It respects the provided context.
func (fl *FileLogger) Close(ctx context.Context) error {
	fl.mu.Lock()
	if fl.closed {
		fl.mu.Unlock()
		return nil
	}
	fl.closed = true
	fl.mu.Unlock()

	// 1. Close channel so the worker finishes reading all remaining alerts
	close(fl.ch)

	// 2. Wait for worker in a channel select block to respect the graceful shutdown context
	done := make(chan struct{})
	go func() {
		fl.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean completion
	case <-ctx.Done():
		// Timeout/Cancellation
		fl.cancel()
		return fmt.Errorf("graceful logger close timed out: %w", ctx.Err())
	}

	// 3. Close the file descriptor safely
	if err := fl.file.Sync(); err != nil {
		_ = fl.file.Close()
		return fmt.Errorf("failed to sync log file before closing: %w", err)
	}

	if err := fl.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	fl.cancel()
	return nil
}
