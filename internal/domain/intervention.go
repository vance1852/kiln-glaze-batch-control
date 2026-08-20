package domain

import (
	"fmt"
	"strings"
	"time"
)

type HealthAlertStatus string

const (
	HealthAlertOpen       HealthAlertStatus = "open"
	HealthAlertInProgress HealthAlertStatus = "in_progress"
	HealthAlertClosed     HealthAlertStatus = "closed"
)

type HealthAlert struct {
	ID              string
	DeploymentJobID string
	Kind            string
	Status          HealthAlertStatus
	Reason          string
	DueAt           time.Time
	ClosedAt        *time.Time
}

func (d HealthAlert) Validate(now time.Time) error {
	if d.DeploymentJobID == "" || strings.TrimSpace(d.Kind) == "" || strings.TrimSpace(d.Reason) == "" {
		return fmt.Errorf("safety_alert fields are required: %w", ErrConflict)
	}
	switch d.Kind {
	case "reassess", "repeat_managed_device", "safety_adjustment", "close_record":
	default:
		return fmt.Errorf("safety_alert kind is invalid: %w", ErrConflict)
	}
	if d.DueAt.Before(now.Add(-24 * time.Hour)) {
		return fmt.Errorf("safety_alert due time is too old: %w", ErrConflict)
	}
	if d.Status == HealthAlertClosed && d.ClosedAt == nil {
		return fmt.Errorf("closed safety_alert needs closed_at: %w", ErrConflict)
	}
	return nil
}

func (d HealthAlert) IsDue(now time.Time) bool {
	return d.Status != HealthAlertClosed && !d.DueAt.After(now)
}
