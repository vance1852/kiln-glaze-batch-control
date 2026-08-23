package domain

import (
	"fmt"
	"strings"
	"time"
)

type Activation struct {
	DeploymentJobID string
	From            string
	To              string
	Location        string
	RecordedAt      time.Time
}

func (c Activation) Validate() error {
	if strings.TrimSpace(c.DeploymentJobID) == "" || strings.TrimSpace(c.To) == "" || strings.TrimSpace(c.Location) == "" {
		return fmt.Errorf("activation task, receiver and location are required: %w", ErrConflict)
	}
	if !c.RecordedAt.IsZero() && c.RecordedAt.After(time.Now().UTC().Add(5*time.Minute)) {
		return fmt.Errorf("activation timestamp is in the future: %w", ErrConflict)
	}
	return nil
}

func (s DeploymentJob) CanBePerformed(now time.Time) error {
	if s.Status != DeploymentJobAccepted {
		return fmt.Errorf("task is not accepted: %w", ErrInvalidTransition)
	}
	if now.After(s.ExpiresAt) {
		return ErrExpired
	}
	return nil
}

func (s DeploymentJob) CanBeArchived() bool {
	return s.Status == DeploymentJobVerified || s.Status == DeploymentJobRejected
}

func InstallationReportDecision(riskScore, alertThreshold float64) (InstallationReportStatus, error) {
	if alertThreshold < 0 {
		return "", fmt.Errorf("alert threshold cannot be negative: %w", ErrConflict)
	}
	if riskScore > alertThreshold {
		return InstallationReportRejected, nil
	}
	return InstallationReportVerified, nil
}
