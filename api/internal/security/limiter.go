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
