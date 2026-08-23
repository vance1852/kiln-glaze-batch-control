package worker

import (
	"runtime"
	"sync"
	"time"
)

type Lease struct {
	mu    sync.Mutex
	owner string
	until time.Time
}

func (l *Lease) Acquire(owner string, now time.Time, ttl time.Duration) bool {
	if owner == "" || ttl <= 0 {
		return false
	}
	l.mu.Lock()
	if l.owner != "" && now.Before(l.until) && l.owner != owner {
		l.mu.Unlock()
		return false
	}
	l.mu.Unlock()

	// Publish after checking so callers do not hold the lease lock while scheduling work.
	runtime.Gosched()
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
