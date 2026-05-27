package core

import (
	"fmt"
	"net"
	"sync"
)

// WhitelistRegistry manages IP addresses and CIDR subnets that are explicitly
// exempt from being blocked by Sananti's threat detection mechanisms.
type WhitelistRegistry struct {
	mu    sync.RWMutex
	ips   map[string]bool
	subnets []*net.IPNet
}

// NewWhitelistRegistry initializes an empty WhitelistRegistry.
func NewWhitelistRegistry() *WhitelistRegistry {
	return &WhitelistRegistry{
		ips: make(map[string]bool),
	}
}

// Add registers a single IP address (e.g., "127.0.0.1", "::1") or a CIDR subnet
// (e.g., "192.168.0.0/16", "10.0.0.0/8") into the whitelisted registry.
func (w *WhitelistRegistry) Add(ipOrCIDR string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Attempt to parse as a CIDR subnet block
	_, subnet, err := net.ParseCIDR(ipOrCIDR)
	if err == nil && subnet != nil {
		w.subnets = append(w.subnets, subnet)
		return nil
	}

	// 2. Attempt to parse as a direct singular IP address
	ip := net.ParseIP(ipOrCIDR)
	if ip != nil {
		w.ips[ip.String()] = true
		return nil
	}

	return fmt.Errorf("invalid IP address or CIDR format: %q", ipOrCIDR)
}

// Contains checks if a given IP address is whitelisted, either directly
// or because it falls inside an authorized CIDR subnet block.
func (w *WhitelistRegistry) Contains(ipStr string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Parse incoming IP string
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 1. Check singular IP exact match
	if w.ips[ip.String()] {
		return true
	}

	// 2. Check CIDR subnets containment
	for _, subnet := range w.subnets {
		if subnet.Contains(ip) {
			return true
		}
	}

	return false
}
