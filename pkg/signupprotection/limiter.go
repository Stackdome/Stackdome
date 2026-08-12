package signupprotection

import (
	"crypto/sha256"
	"errors"
	"net/netip"
	"strings"
	"sync"
	"time"
)

// ThrottleSpec configures one bounded fixed-window attempt store.
type ThrottleSpec struct {
	MaxTrackedKeys int
	MaxAttempts    int
	Window         time.Duration
	Now            func() time.Time
}

type fixedWindowEntry struct {
	windowStart time.Time
	count       int
}

type fixedWindowLimiter[Key comparable] struct {
	mutex          sync.Mutex
	entries        map[Key]fixedWindowEntry
	maxTrackedKeys int
	maxAttempts    int
	window         time.Duration
	now            func() time.Time
	nextExpiry     time.Time
}

func newFixedWindowLimiter[Key comparable](spec ThrottleSpec) (*fixedWindowLimiter[Key], error) {
	if spec.MaxTrackedKeys <= 0 {
		return nil, errors.New("throttle max tracked keys must be positive")
	}
	if spec.MaxAttempts <= 0 {
		return nil, errors.New("throttle max attempts must be positive")
	}
	if spec.Window <= 0 {
		return nil, errors.New("throttle window must be positive")
	}
	if spec.Now == nil {
		return nil, errors.New("throttle clock is required")
	}

	return &fixedWindowLimiter[Key]{
		entries:        make(map[Key]fixedWindowEntry, spec.MaxTrackedKeys),
		maxTrackedKeys: spec.MaxTrackedKeys,
		maxAttempts:    spec.MaxAttempts,
		window:         spec.Window,
		now:            spec.Now,
	}, nil
}

func (l *fixedWindowLimiter[Key]) Allow(key Key) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := l.now()
	if entry, found := l.entries[key]; found {
		if !now.Before(entry.windowStart.Add(l.window)) {
			entry.windowStart = now
			entry.count = 1
			l.entries[key] = entry
			return true
		}
		if entry.count >= l.maxAttempts {
			return false
		}

		entry.count++
		l.entries[key] = entry
		return true
	}

	if len(l.entries) >= l.maxTrackedKeys {
		if now.Before(l.nextExpiry) {
			return false
		}
		l.removeExpiredEntries(now)
		if len(l.entries) >= l.maxTrackedKeys {
			return false
		}
	}

	l.entries[key] = fixedWindowEntry{windowStart: now, count: 1}
	l.trackNextExpiry(now.Add(l.window))
	return true
}

func (l *fixedWindowLimiter[Key]) removeExpiredEntries(now time.Time) {
	l.nextExpiry = time.Time{}
	for key, entry := range l.entries {
		expiresAt := entry.windowStart.Add(l.window)
		if !now.Before(expiresAt) {
			delete(l.entries, key)
			continue
		}
		l.trackNextExpiry(expiresAt)
	}
}

func (l *fixedWindowLimiter[Key]) trackNextExpiry(expiresAt time.Time) {
	if l.nextExpiry.IsZero() || expiresAt.Before(l.nextExpiry) {
		l.nextExpiry = expiresAt
	}
}

type ipLimiter struct {
	store *fixedWindowLimiter[netip.Addr]
}

func newIPLimiter(spec ThrottleSpec) (*ipLimiter, error) {
	store, err := newFixedWindowLimiter[netip.Addr](spec)
	if err != nil {
		return nil, err
	}
	return &ipLimiter{store: store}, nil
}

func (l *ipLimiter) Allow(clientIP netip.Addr) bool {
	return l.store.Allow(clientIP.Unmap())
}

type emailLimiter struct {
	store *fixedWindowLimiter[[sha256.Size]byte]
}

func newEmailLimiter(spec ThrottleSpec) (*emailLimiter, error) {
	store, err := newFixedWindowLimiter[[sha256.Size]byte](spec)
	if err != nil {
		return nil, err
	}
	return &emailLimiter{store: store}, nil
}

func (l *emailLimiter) Allow(email string) bool {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	return l.store.Allow(sha256.Sum256([]byte(normalizedEmail)))
}
