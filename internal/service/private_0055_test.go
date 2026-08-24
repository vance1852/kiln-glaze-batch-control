package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/repository"
)

type privateActivationRepo0055 struct { repository.Repository }
func (r *privateActivationRepo0055) ListActivation(ctx context.Context, taskID string) ([]repository.ActivationRecord, error) {
  if !errors.Is(ctx.Err(), context.Canceled) { return nil, errors.New("activation lookup outlived request") }
  return nil, context.Canceled
}
func TestActivationLookupHonorsRequestCancellation(t *testing.T) {
  ctx, cancel := context.WithCancel(context.Background()); cancel()
  _, err := New(&privateActivationRepo0055{}).VerifyActivation(ctx, "activation-55")
  if !errors.Is(err, context.Canceled) { t.Fatalf("err=%v", err) }
}
