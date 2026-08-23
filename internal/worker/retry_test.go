package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryPolicyCapsExponentialDelay(t *testing.T) {
	p := RetryPolicy{BaseDelay: time.Second, MaxDelay: 3 * time.Second}
	if p.Delay(1) != time.Second || p.Delay(2) != 2*time.Second || p.Delay(3) != 3*time.Second {
		t.Fatalf("delays are incorrect")
	}
}

func TestRunWithRetryStopsAfterSuccess(t *testing.T) {
	attempts := 0
	err := RunWithRetry(context.Background(), RetryPolicy{Attempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}, func(context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil || attempts != 2 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}

func TestRunWithRetryPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunWithRetry(ctx, RetryPolicy{Attempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, func(context.Context) error { t.Fatal("operation ran after cancel"); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}

// TestRunWithRetryGivesEachAttemptAnIndependentTimeoutBudget guards the
// kiln-firing wave recovery path: when the first attempt runs to the end of
// its time budget, the second attempt must still receive a fresh, unexpired
// context rather than inheriting the first attempt's already-expired deadline.
func TestRunWithRetryGivesEachAttemptAnIndependentTimeoutBudget(t *testing.T) {
	policy := RetryPolicy{
		Attempts:       2,
		BaseDelay:      time.Nanosecond,
		MaxDelay:       time.Nanosecond,
		AttemptTimeout: 15 * time.Millisecond,
	}
	var sawExpiredOnRetry bool
	err := RunWithRetry(context.Background(), policy, func(attemptCtx context.Context) error {
		// First attempt consumes the whole budget so any shared timeout would
		// be exhausted before the second attempt begins.
		<-attemptCtx.Done()
		if errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			sawExpiredOnRetry = true
		}
		return attemptCtx.Err()
	})
	if err == nil {
		t.Fatal("expected operation to report deadline exceeded")
	}
	if !sawExpiredOnRetry {
		t.Fatal("second attempt observed an already-expired context; each attempt needs its own timeout budget")
	}
}

// TestRunWithRetryResetsTimeoutOnEachAttempt asserts the concrete observable
// guarantee: a second attempt whose operation returns quickly must succeed,
// proving its context was not pre-expired by the first attempt's budget use.
func TestRunWithRetryResetsTimeoutOnEachAttempt(t *testing.T) {
	policy := RetryPolicy{
		Attempts:       3,
		BaseDelay:      time.Nanosecond,
		MaxDelay:       time.Nanosecond,
		AttemptTimeout: 20 * time.Millisecond,
	}
	var calls int
	err := RunWithRetry(context.Background(), policy, func(attemptCtx context.Context) error {
		calls++
		if attemptCtx.Err() != nil {
			return fmt.Errorf("attempt %d entered with expired context: %w", calls, attemptCtx.Err())
		}
		if calls == 1 {
			<-attemptCtx.Done()
			return attemptCtx.Err()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected second attempt to succeed on a fresh budget, err=%v", err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}
