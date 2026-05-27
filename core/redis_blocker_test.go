package core

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisBlocker_Operations(t *testing.T) {
	// Attempt connection to a local Redis server
	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379",
		DialTimeout: 500 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Skipping RedisBlocker integration test: local Redis server is offline on port 6379")
		return
	}
	defer rdb.Close()

	// Clear test keys before starting
	testIP := "198.51.100.123"
	testKey := "sananti:blocked:" + testIP
	_ = rdb.Del(context.Background(), testKey).Err()
	defer func() {
		_ = rdb.Del(context.Background(), testKey).Err()
	}()

	// Initialize Redis blocker with a small block time (TTL)
	blocker := NewRedisBlocker(rdb, 10*time.Second)

	// 1. Initial State Check
	blocked, _, err := blocker.IsBlocked(testIP)
	if err != nil {
		t.Fatalf("unexpected error on initial IsBlocked check: %v", err)
	}
	if blocked {
		t.Errorf("expected IP %s to not be blocked initially", testIP)
	}

	// 2. Block the IP
	err = blocker.BlockIP(testIP, "decoy triggered")
	if err != nil {
		t.Fatalf("failed to block IP: %v", err)
	}

	blocked, reason, err := blocker.IsBlocked(testIP)
	if err != nil {
		t.Fatalf("unexpected error on second IsBlocked check: %v", err)
	}
	if !blocked {
		t.Errorf("expected IP %s to be blocked after BlockIP", testIP)
	}
	if reason != "decoy triggered" {
		t.Errorf("expected reason 'decoy triggered', got %q", reason)
	}

	// 3. Increment Attempts on second trigger
	err = blocker.BlockIP(testIP, "second trigger")
	if err != nil {
		t.Fatalf("failed to re-block IP: %v", err)
	}

	snapshot := blocker.GetBlockedIPs()
	info, exists := snapshot[testIP]
	if !exists {
		t.Fatalf("expected test IP to exist in Redis snapshot")
	}
	if info.Attempts != 2 {
		t.Errorf("expected attempts to be 2, got %d", info.Attempts)
	}
	if info.Reason != "second trigger" {
		t.Errorf("expected updated reason, got %q", info.Reason)
	}
}
