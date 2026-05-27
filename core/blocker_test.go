package core

import (
	"sync"
	"testing"
	"time"
)

func TestMemoryBlocker_Operations(t *testing.T) {
	// Initialize blocker with 2-second default TTL
	mb := NewMemoryBlocker(2 * time.Second)
	defer mb.Close()

	ip := "192.0.2.1"

	// 1. Initial State Check
	blocked, _, err := mb.IsBlocked(ip)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blocked {
		t.Errorf("expected IP %s to not be blocked initially", ip)
	}

	// 2. Block the IP
	err = mb.BlockIP(ip, "Decoy triggered")
	if err != nil {
		t.Fatalf("unexpected error on BlockIP: %v", err)
	}

	blocked, reason, err := mb.IsBlocked(ip)
	if err != nil {
		t.Fatalf("unexpected error on second check: %v", err)
	}
	if !blocked {
		t.Errorf("expected IP %s to be blocked", ip)
	}
	if reason != "Decoy triggered" {
		t.Errorf("expected reason 'Decoy triggered', got %q", reason)
	}

	// 3. Increment Attempts
	err = mb.BlockIP(ip, "Decoy triggered again")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snapshot := mb.GetBlockedIPs()
	info, exists := snapshot[ip]
	if !exists {
		t.Fatalf("expected blocked IP %s to exist in snapshot", ip)
	}
	if info.Attempts != 2 {
		t.Errorf("expected attempts to be 2, got %d", info.Attempts)
	}

	// 4. Test TTL Lazy Unblocking
	// Sleep for 2.1 seconds to let the TTL expire
	time.Sleep(2100 * time.Millisecond)

	blockedExpired, _, err := mb.IsBlocked(ip)
	if err != nil {
		t.Fatalf("unexpected error on expired check: %v", err)
	}
	if blockedExpired {
		t.Error("expected IP to be automatically unblocked after TTL expiration")
	}
}

func TestMemoryBlocker_CIDRWhitelisting(t *testing.T) {
	mb := NewMemoryBlocker(1 * time.Hour)
	defer mb.Close()

	// Whitelist an administrative CIDR subnet block
	err := mb.GetWhitelist().Add("192.168.1.0/24")
	if err != nil {
		t.Fatalf("failed to add CIDR: %v", err)
	}

	whitelistedIP := "192.168.1.100"
	attackerIP := "203.0.113.50"

	// Attempt to block whitelisted IP
	err = mb.BlockIP(whitelistedIP, "decoy trigger")
	if err != nil {
		t.Fatalf("unexpected block error: %v", err)
	}

	blocked, _, _ := mb.IsBlocked(whitelistedIP)
	if blocked {
		t.Error("whitelisted IP was blocked, whitelisting failed to exempt IP")
	}

	// Attempt to block normal non-whitelisted IP
	err = mb.BlockIP(attackerIP, "decoy trigger")
	if err != nil {
		t.Fatalf("unexpected block error: %v", err)
	}

	blockedAttacker, _, _ := mb.IsBlocked(attackerIP)
	if !blockedAttacker {
		t.Error("non-whitelisted IP was not blocked")
	}
}

func TestMemoryBlocker_Concurrency(t *testing.T) {
	mb := NewMemoryBlocker(1 * time.Hour)
	defer mb.Close()

	var wg sync.WaitGroup
	numGoroutines := 100
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = mb.BlockIP("192.0.2.1", "concurrency test")
		}()

		go func() {
			defer wg.Done()
			_, _, _ = mb.IsBlocked("192.0.2.1")
		}()
	}

	wg.Wait()
}
