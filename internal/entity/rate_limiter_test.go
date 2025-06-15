package entity

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	ip := "192.168.1.1"
	token := "test-token"

	limiter := NewRateLimiter(ip, token)

	if limiter.IP != ip {
		t.Errorf("Expected IP %s, got %s", ip, limiter.IP)
	}

	if limiter.Token != token {
		t.Errorf("Expected Token %s, got %s", token, limiter.Token)
	}

	if limiter.Requests != 0 {
		t.Errorf("Expected Requests 0, got %d", limiter.Requests)
	}

	if !limiter.LastRequest.IsZero() {
		t.Error("Expected LastRequest to be zero time")
	}

	if !limiter.BlockedUntil.IsZero() {
		t.Error("Expected BlockedUntil to be zero time")
	}

	if limiter.Blocked {
		t.Error("Expected Blocked to be false")
	}
}

func TestIsBlocked(t *testing.T) {
	tests := []struct {
		name            string
		limiter         *RateLimiter
		expectedBlocked bool
	}{
		{
			name: "Not blocked",
			limiter: &RateLimiter{
				Blocked:      false,
				BlockedUntil: time.Now().Add(time.Hour),
			},
			expectedBlocked: false,
		},
		{
			name: "Blocked and not expired",
			limiter: &RateLimiter{
				Blocked:      true,
				BlockedUntil: time.Now().Add(time.Hour),
			},
			expectedBlocked: true,
		},
		{
			name: "Blocked but expired",
			limiter: &RateLimiter{
				Blocked:      true,
				BlockedUntil: time.Now().Add(-time.Hour),
			},
			expectedBlocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked := tt.limiter.IsBlocked()
			if blocked != tt.expectedBlocked {
				t.Errorf("Expected IsBlocked() to be %v, got %v", tt.expectedBlocked, blocked)
			}
		})
	}
}

func TestBlockAndUnblock(t *testing.T) {
	limiter := NewRateLimiter("192.168.1.1", "")
	duration := time.Hour

	// Test Block
	limiter.Block(duration)
	if !limiter.Blocked {
		t.Error("Expected Blocked to be true after Block()")
	}

	if limiter.BlockedUntil.Before(time.Now()) {
		t.Error("Expected BlockedUntil to be in the future")
	}

	// Test Unblock
	limiter.Unblock()
	if limiter.Blocked {
		t.Error("Expected Blocked to be false after Unblock()")
	}

	if !limiter.BlockedUntil.IsZero() {
		t.Error("Expected BlockedUntil to be zero time after Unblock()")
	}
}
