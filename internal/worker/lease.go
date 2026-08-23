package worker

import (
	"sync"
	"time"
)

type Lease struct {
	mu    sync.Mutex
	owner string
	until time.Time
}

// Acquire grants ownership to at most one caller at a time. The vacancy check
// and the ownership write happen inside a single critical section so that two
// workers contending for an idle lease cannot both observe the vacancy and
// both succeed; the loser is rejected outright.
func (l *Lease) Acquire(owner string, now time.Time, ttl time.Duration) bool {
	if owner == "" || ttl <= 0 {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owner != "" && now.Before(l.until) && l.owner != owner {
		return false
	}
	l.owner, l.until = owner, now.Add(ttl)
	return true
}

func (l *Lease) Release(owner string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owner != owner {
		return false
	}
	l.owner, l.until = "", time.Time{}
	return true
}

func (l *Lease) HeldBy(owner string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.owner == owner && now.Before(l.until)
}
