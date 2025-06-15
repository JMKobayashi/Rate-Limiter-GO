package entity

import "time"

type RateLimiter struct {
	IP           string
	Token        string
	Requests     int64
	LastRequest  time.Time
	BlockedUntil time.Time
	Blocked      bool
}

func NewRateLimiter(ip, token string) *RateLimiter {
	return &RateLimiter{
		IP:    ip,
		Token: token,
	}
}

func (r *RateLimiter) IsBlocked() bool {
	return r.Blocked && time.Now().Before(r.BlockedUntil)
}

func (r *RateLimiter) Block(duration time.Duration) {
	r.Blocked = true
	r.BlockedUntil = time.Now().Add(duration)
}

func (r *RateLimiter) Unblock() {
	r.Blocked = false
	r.BlockedUntil = time.Time{}
}
