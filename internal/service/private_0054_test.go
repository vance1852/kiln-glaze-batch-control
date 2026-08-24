package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/repository"
)

type privateActivationRepo0054 struct { repository.Repository }
func (r *privateActivationRepo0054) ListActivation(ctx context.Context, taskID string) ([]repository.ActivationRecord, error) {
  if !errors.Is(ctx.Err(), context.Canceled) { return nil, errors.New("lookup context was not cancelled") }
  return nil, context.Canceled
}
func TestVerifyActivationPropagatesCancellation(t *testing.T) {
  ctx, cancel := context.WithCancel(context.Background()); cancel()
  _, err := New(&privateActivationRepo0054{}).VerifyActivation(ctx, "task-54")
  if !errors.Is(err, context.Canceled) { t.Fatalf("err=%v", err) }
}
