package security

import (
	"sync"
	"time"
)

type loginAttempt struct {
	count       int
	lockedUntil time.Time
	lastSeen    time.Time
}

type LoginLimiter struct {
	mu          sync.Mutex
	attempts    map[string]loginAttempt
	maxAttempts int
	lockout     time.Duration
}

func NewLoginLimiter(maxAttempts int, lockout time.Duration) *LoginLimiter {
	return &LoginLimiter{
		attempts:    make(map[string]loginAttempt),
		maxAttempts: maxAttempts,
		lockout:     lockout,
	}
}

func (l *LoginLimiter) Allowed(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt := l.attempts[key]
	if attempt.lockedUntil.After(now) {
		return false
	}
	if !attempt.lockedUntil.IsZero() {
		delete(l.attempts, key)
	}
	return true
}

func (l *LoginLimiter) Failure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	attempt := l.attempts[key]
	attempt.count++
	attempt.lastSeen = now
	if attempt.count >= l.maxAttempts {
		attempt.lockedUntil = now.Add(l.lockout)
	}
	l.attempts[key] = attempt
	if len(l.attempts) > 10_000 {
		for attemptKey, candidate := range l.attempts {
			if now.Sub(candidate.lastSeen) > l.lockout*2 {
				delete(l.attempts, attemptKey)
			}
		}
	}
}

func (l *LoginLimiter) Success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

type RateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

func (l *RateLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	recent := l.recentHits(key, now)
	if len(recent) >= l.limit {
		l.hits[key] = recent
		return false
	}
	l.hits[key] = append(recent, now)
	if len(l.hits) > 10_000 {
		for hitKey, hits := range l.hits {
			if !hits[len(hits)-1].After(l.windowStart(now)) {
				delete(l.hits, hitKey)
			}
		}
	}
	return true
}

func (l *RateLimiter) recentHits(key string, now time.Time) []time.Time {
	start := l.windowStart(now)
	hits := l.hits[key]
	recent := hits[:0]
	for _, hit := range hits {
		if hit.After(start) {
			recent = append(recent, hit)
		}
	}
	return recent
}

func (l *RateLimiter) windowStart(now time.Time) time.Time {
	return now.Add(-l.window)
}
