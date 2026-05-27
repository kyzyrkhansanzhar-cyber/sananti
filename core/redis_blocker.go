package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBlocker implements the Blocker interface utilizing a live Redis database.
// This permits multi-server distributed synchronization and temporary TTL-based blocking.
type RedisBlocker struct {
	client     *redis.Client
	defaultTTL time.Duration
	prefix     string
	whitelist  *WhitelistRegistry
}

// NewRedisBlocker initializes a new distributed Blocker backed by Redis.
func NewRedisBlocker(client *redis.Client, defaultTTL time.Duration) *RedisBlocker {
	return &RedisBlocker{
		client:     client,
		defaultTTL: defaultTTL,
		prefix:     "sananti:blocked:",
		whitelist:  NewWhitelistRegistry(),
	}
}

// GetWhitelist returns the safe IP/CIDR whitelist registry.
func (r *RedisBlocker) GetWhitelist() *WhitelistRegistry {
	return r.whitelist
}

// BlockIP blacklists an IP address inside Redis utilizing the default TTL.
func (r *RedisBlocker) BlockIP(ip string, reason string) error {
	return r.BlockIPWithTTL(ip, reason, r.defaultTTL)
}

// BlockIPWithTTL blacklists an IP address inside Redis with a custom TTL duration.
// If the IP is whitelisted, the request is ignored safely.
func (r *RedisBlocker) BlockIPWithTTL(ip string, reason string, ttl time.Duration) error {
	// Ensure IP is not whitelisted
	if r.whitelist.Contains(ip) {
		return nil
	}

	ctx := context.Background()
	key := r.prefix + ip

	// Check if already blocked
	val, err := r.client.Get(ctx, key).Result()
	var info BlockInfo

	expiresAt := time.Now().Add(ttl)

	if err == nil {
		// Key exists, deserialize and update
		if err := json.Unmarshal([]byte(val), &info); err == nil {
			info.Attempts++
			info.BlockedAt = time.Now()
			info.ExpiresAt = expiresAt
			info.Reason = reason
		} else {
			// Fallback if deserialization fails
			info = BlockInfo{BlockedAt: time.Now(), ExpiresAt: expiresAt, Reason: reason, Attempts: 1}
		}
	} else if errors.Is(err, redis.Nil) {
		// New blocked IP entry
		info = BlockInfo{
			BlockedAt: time.Now(),
			ExpiresAt: expiresAt,
			Reason:    reason,
			Attempts:  1,
		}
	} else {
		return fmt.Errorf("failed to query Redis on BlockIP: %w", err)
	}

	// Serialize updated record
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal block info for Redis: %w", err)
	}

	// Save in Redis with TTL expiration
	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to write blocked IP to Redis: %w", err)
	}

	return nil
}

// IsBlocked checks if an IP address is blacklisted in Redis.
func (r *RedisBlocker) IsBlocked(ip string) (bool, string, error) {
	ctx := context.Background()
	key := r.prefix + ip

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, "", nil
	} else if err != nil {
		return false, "", fmt.Errorf("failed to check Redis blacklist: %w", err)
	}

	var info BlockInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return true, "Blocked (corrupt metadata)", nil
	}

	// Verify expiration just in case Redis TTL hasn't evicted it yet
	if time.Now().After(info.ExpiresAt) {
		_ = r.UnblockIP(ip)
		return false, "", nil
	}

	return true, info.Reason, nil
}

// UnblockIP manually removes an IP from Redis.
func (r *RedisBlocker) UnblockIP(ip string) error {
	ctx := context.Background()
	key := r.prefix + ip
	return r.client.Del(ctx, key).Err()
}

// GetBlockedIPs retrieves all currently active blacklisted IPs by scanning keys.
func (r *RedisBlocker) GetBlockedIPs() map[string]BlockInfo {
	ctx := context.Background()
	snapshot := make(map[string]BlockInfo)

	// Scan keys matching our prefix
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, r.prefix+"*", 100).Result()
		if err != nil {
			_, _ = fmt.Printf("[Sananti Redis Blocker Error] Keys Scan failed: %v\n", err)
			return snapshot
		}

		for _, key := range keys {
			val, err := r.client.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var info BlockInfo
			if err := json.Unmarshal([]byte(val), &info); err == nil {
				// Only return non-expired entries
				if time.Now().Before(info.ExpiresAt) {
					ip := strings.TrimPrefix(key, r.prefix)
					snapshot[ip] = info
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return snapshot
}
