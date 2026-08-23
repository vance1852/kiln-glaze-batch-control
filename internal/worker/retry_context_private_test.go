package worker

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRetryAttemptTimeoutUsesFreshContext(t *testing.T) {
	attempts := 0
	policy := RetryPolicy{
		Attempts:       2,
		BaseDelay:      time.Nanosecond,
		MaxDelay:       time.Nanosecond,
		AttemptTimeout: 20 * time.Millisecond,
	}

	err := RunWithRetry(t.Context(), policy, func(ctx context.Context) error {
		attempts++
		if attempts == 1 {
			<-ctx.Done()
			return ctx.Err()
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry received expired attempt context: %w", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("firing-wave retry did not recover after one attempt timeout: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempt count = %d, want 2", attempts)
	}
}
