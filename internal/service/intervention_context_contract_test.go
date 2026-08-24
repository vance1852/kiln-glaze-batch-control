package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"firmware-rollout-control/internal/repository"
)

type cancellationRepository struct {
	repository.Repository
}

func (cancellationRepository) DueHealthAlerts(ctx context.Context, _ time.Time, _ int) ([]repository.HealthAlertInput, error) {
	return nil, ctx.Err()
}

func (cancellationRepository) DueHealthAlertsDetached(context.Context, time.Time, int) ([]repository.HealthAlertInput, error) {
	return nil, nil
}

func TestDueHealthAlertQueryPreservesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := New(cancellationRepository{}).DueHealthAlerts(ctx, time.Now().UTC(), 10)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
