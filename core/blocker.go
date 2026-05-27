package core

import (
	"context"
	"sync"
	"time"
)

// Blocker defines the behavioral contract for handling client IP blacklists.
type Blocker interface {
	// BlockIP blacklists an IP address with a default TTL duration.
	BlockIP(ip string, reason string) error

	// BlockIPWithTTL blacklists an IP address with a specific custom TTL duration.
	BlockIPWithTTL(ip string, reason string, ttl time.Duration) error

	// IsBlocked checks if an IP is blocked. Returns (isBlocked, reason, error).
	IsBlocked(ip string) (bool, string, error)

	// UnblockIP manually removes an IP from the blacklist.
	UnblockIP(ip string) error

	// GetBlockedIPs returns a snapshot copy of all currently active blacklisted IPs.
	GetBlockedIPs() map[string]BlockInfo

	// GetWhitelist returns the safe IP whitelist registry.
	GetWhitelist() *WhitelistRegistry
}

// MemoryBlocker implements the Blocker interface utilizing a thread-safe
// in-memory map featuring CIDR whitelisting and active background/lazy TTL unblocking.
type MemoryBlocker struct {
	mu         sync.RWMutex
	blocked    map[string]BlockInfo
	whitelist  *WhitelistRegistry
	defaultTTL time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewMemoryBlocker initializes a new in-memory blocker with TTL and active background cleaning.
func NewMemoryBlocker(defaultTTL time.Duration) *MemoryBlocker {
	ctx, cancel := context.WithCancel(context.Background())
	mb := &MemoryBlocker{
		blocked:    make(map[string]BlockInfo),
		whitelist:  NewWhitelistRegistry(),
		defaultTTL: defaultTTL,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start active background cleanup worker running every 10 seconds to keep memory minimal
	mb.wg.Add(1)
	go mb.cleanupWorker(10 * time.Second)

	return mb
}

// GetWhitelist returns the whitelisted IP/CIDR registry.
func (m *MemoryBlocker) GetWhitelist() *WhitelistRegistry {
	return m.whitelist
}

// BlockIP adds an IP to the blacklist utilizing the default configured TTL.
func (m *MemoryBlocker) BlockIP(ip string, reason string) error {
	return m.BlockIPWithTTL(ip, reason, m.defaultTTL)
}

// BlockIPWithTTL adds an IP to the blacklist with a custom expiration duration.
// If the IP is whitelisted, the request is safely ignored.
func (m *MemoryBlocker) BlockIPWithTTL(ip string, reason string, ttl time.Duration) error {
	// 1. Ensure IP is not whitelisted
	if m.whitelist.Contains(ip) {
		// Log or ignore safely
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt := time.Now().Add(ttl)
	info, exists := m.blocked[ip]
	if exists {
		// Increment attempts and update block metadata
		info.Attempts++
		info.BlockedAt = time.Now()
		info.ExpiresAt = expiresAt
		info.Reason = reason
		m.blocked[ip] = info
	} else {
		m.blocked[ip] = BlockInfo{
			BlockedAt: time.Now(),
			ExpiresAt: expiresAt,
			Reason:    reason,
			Attempts:  1,
		}
	}

	return nil
}

// IsBlocked checks if an IP is blocked. If an IP has expired, it triggers a
// lazy unblock deletion and returns false.
func (m *MemoryBlocker) IsBlocked(ip string) (bool, string, error) {
	m.mu.RLock()
	info, exists := m.blocked[ip]
	m.mu.RUnlock()

	if !exists {
		return false, "", nil
	}

	// Check if block duration has expired
	if time.Now().After(info.ExpiresAt) {
		// Trigger a lazy deletion unblock
		m.mu.Lock()
		delete(m.blocked, ip)
		m.mu.Unlock()
		return false, "", nil
	}

	return true, info.Reason, nil
}

// UnblockIP manually removes an IP from the blacklist map.
func (m *MemoryBlocker) UnblockIP(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.blocked, ip)
	return nil
}

// GetBlockedIPs returns a deep copy of all active blocked IPs, excluding expired records.
func (m *MemoryBlocker) GetBlockedIPs() map[string]BlockInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	snapshot := make(map[string]BlockInfo)

	for ip, info := range m.blocked {
		if now.Before(info.ExpiresAt) {
			snapshot[ip] = info
		}
	}

	return snapshot
}

// cleanupWorker regularly purges expired block keys in the background.
func (m *MemoryBlocker) cleanupWorker(interval time.Duration) {
	defer m.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for ip, info := range m.blocked {
				if now.After(info.ExpiresAt) {
					delete(m.blocked, ip)
				}
			}
			m.mu.Unlock()
		case <-m.ctx.Done():
			return
		}
	}
}

// Close cancels background routines safely.
func (m *MemoryBlocker) Close() error {
	m.cancel()
	m.wg.Wait()
	return nil
}
