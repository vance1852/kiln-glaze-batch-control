package rollout

import (
	"context"
	"errors"
	"testing"
)

func TestOnceStepRetriesAfterFailedExecution(t *testing.T) {
	var calls int
	step := NewOnceStep(func(context.Context, string) error {
		calls++
		if calls == 1 {
			return errors.New("temporary kiln controller outage")
		}
		return nil
	})
	if err := step.Run(context.Background(), "wave-14"); err == nil {
		t.Fatal("first execution unexpectedly succeeded")
	}
	if err := step.Run(context.Background(), "wave-14"); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d, failed work item was permanently suppressed", calls)
	}
}
