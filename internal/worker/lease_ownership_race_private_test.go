package worker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLeaseConcurrentAcquireHasSingleOwner(t *testing.T) {
	for round := 0; round < 32; round++ {
		lease := &Lease{}
		start := make(chan struct{})
		var owners atomic.Int64
		var workers sync.WaitGroup
		for candidate := 0; candidate < 8; candidate++ {
			candidate := candidate
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				if lease.Acquire(fmt.Sprintf("kiln-worker-%d", candidate), time.Unix(100, 0), time.Minute) {
					owners.Add(1)
				}
			}()
		}
		close(start)
		workers.Wait()
		if got := owners.Load(); got != 1 {
			t.Fatalf("round %d granted the firing-wave lease to %d workers, want exactly one", round, got)
		}
	}
}
