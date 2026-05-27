package antifraud

import "testing"

func TestResolveCountry(t *testing.T) {
	tests := []struct {
		ip       string
		expected string
	}{
		{"82.200.3.5", "KZ"},
		{"198.51.100.5", "US"},
		{"91.198.174.192", "DE"},
		{"185.190.140.2", "UK"},
		{"127.0.0.1", "KZ"}, // loopback fallback
		{"127.0.0.1:443", "KZ"}, // strips port successfully
		{"198.51.100.5:8080", "US"}, // strips port successfully
		{"invalid_ip_format", "KZ"}, // invalid fallback
	}

	for _, tt := range tests {
		actual := ResolveCountry(tt.ip)
		if actual != tt.expected {
			t.Errorf("ResolveCountry(%q) = %q; want %q", tt.ip, actual, tt.expected)
		}
	}
}
