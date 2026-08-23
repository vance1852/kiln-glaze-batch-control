package rollout

import (
	"context"
	"errors"
	"testing"
)

func TestOnceStepRetriesAfterTransientFailure(t *testing.T) {
	attempts := 0
	step := NewOnceStep(func(_ context.Context, _ string) error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary kiln controller failure")
		}
		return nil
	})
	if err := step.Run(t.Context(), "wave-0011"); err == nil {
		t.Fatal("first firing step unexpectedly succeeded")
	}
	if err := step.Run(t.Context(), "wave-0011"); err != nil {
		t.Fatalf("retry did not execute: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("executor attempts=%d, want 2", attempts)
	}
}
