package domain

import (
	"fmt"
	"time"
)

type RolloutCampaignStatus string

const (
	RolloutCampaignDraft     RolloutCampaignStatus = "draft"
	RolloutCampaignScheduled RolloutCampaignStatus = "scheduled"
	RolloutCampaignActive    RolloutCampaignStatus = "active"
	RolloutCampaignClosed    RolloutCampaignStatus = "closed"
)

func (s RolloutCampaignStatus) CanMoveTo(next RolloutCampaignStatus) bool {
	switch s {
	case RolloutCampaignDraft:
		return next == RolloutCampaignScheduled
	case RolloutCampaignScheduled:
		return next == RolloutCampaignActive || next == RolloutCampaignClosed
	case RolloutCampaignActive:
		return next == RolloutCampaignClosed
	default:
		return false
	}
}

type DeploymentJobStatus string

const (
	DeploymentJobQueued            DeploymentJobStatus = "queued"
	DeploymentJobCompleted         DeploymentJobStatus = "completed"
	DeploymentJobActivationPending DeploymentJobStatus = "activation_pending"
	DeploymentJobAccepted          DeploymentJobStatus = "accepted"
	DeploymentJobInProgress        DeploymentJobStatus = "in_progress"
	DeploymentJobVerified          DeploymentJobStatus = "verified"
	DeploymentJobRejected          DeploymentJobStatus = "rejected"
	DeploymentJobArchived          DeploymentJobStatus = "archived"
)

func (s DeploymentJobStatus) CanMoveTo(next DeploymentJobStatus) bool {
	switch s {
	case DeploymentJobQueued:
		return next == DeploymentJobCompleted
	case DeploymentJobCompleted:
		return next == DeploymentJobActivationPending
	case DeploymentJobActivationPending:
		return next == DeploymentJobAccepted
	case DeploymentJobAccepted:
		return next == DeploymentJobInProgress
	case DeploymentJobInProgress:
		return next == DeploymentJobVerified || next == DeploymentJobRejected
	case DeploymentJobRejected:
		return next == DeploymentJobArchived
	case DeploymentJobVerified:
		return next == DeploymentJobArchived
	default:
		return false
	}
}

type RolloutCampaign struct {
	ID        string                `json:"id"`
	Code      string                `json:"code"`
	Name      string                `json:"name"`
	Status    RolloutCampaignStatus `json:"status"`
	Timezone  string                `json:"timezone"`
	StartsAt  time.Time             `json:"starts_at"`
	EndsAt    time.Time             `json:"ends_at"`
	Version   int64                 `json:"version"`
	CreatedBy string                `json:"created_by"`
}

func (p RolloutCampaign) ValidateWindow(now time.Time) error {
	if p.EndsAt.Before(p.StartsAt) || p.EndsAt.Equal(p.StartsAt) {
		return fmt.Errorf("rollout_campaign end must be after start: %w", ErrConflict)
	}
	if p.Timezone == "" {
		return fmt.Errorf("timezone is required: %w", ErrConflict)
	}
	if now.After(p.EndsAt) && p.Status != RolloutCampaignClosed {
		return fmt.Errorf("rollout_campaign window has elapsed: %w", ErrExpired)
	}
	return nil
}

type DeploymentJob struct {
	ID                string              `json:"id"`
	RolloutCampaignID string              `json:"rollout_campaign_id"`
	ManagedDeviceID   string              `json:"managed_device_id"`
	TaskCode          string              `json:"task_code"`
	Status            DeploymentJobStatus `json:"status"`
	CompletedAt       *time.Time          `json:"completed_at,omitempty"`
	AcceptedAt        *time.Time          `json:"accepted_at,omitempty"`
	ExpiresAt         time.Time           `json:"expires_at"`
	Version           int64               `json:"version"`
}

func EligibleForExecution(status DeploymentJobStatus, expiresAt, now time.Time) bool {
	if status != DeploymentJobAccepted {
		return false
	}
	if expiresAt.Before(now) {
		return false
	}
	return true
}

func (s DeploymentJob) Move(next DeploymentJobStatus, now time.Time) (DeploymentJob, error) {
	if !s.Status.CanMoveTo(next) {
		return DeploymentJob{}, fmt.Errorf("%s -> %s: %w", s.Status, next, ErrInvalidTransition)
	}
	if now.After(s.ExpiresAt) && next != DeploymentJobArchived {
		return DeploymentJob{}, fmt.Errorf("task %s expired: %w", s.TaskCode, ErrExpired)
	}
	s.Status = next
	s.Version++
	if next == DeploymentJobCompleted {
		s.CompletedAt = &now
	}
	if next == DeploymentJobAccepted {
		s.AcceptedAt = &now
	}
	return s, nil
}

type RolloutWaveStatus string

const (
	RolloutWaveQueued    RolloutWaveStatus = "queued"
	RolloutWaveRunning   RolloutWaveStatus = "running"
	RolloutWaveCompleted RolloutWaveStatus = "completed"
	RolloutWaveCancelled RolloutWaveStatus = "cancelled"
)

func (s RolloutWaveStatus) CanMoveTo(next RolloutWaveStatus) bool {
	switch s {
	case RolloutWaveQueued:
		return next == RolloutWaveRunning || next == RolloutWaveCancelled
	case RolloutWaveRunning:
		return next == RolloutWaveCompleted || next == RolloutWaveCancelled
	default:
		return false
	}
}

type InstallationReportStatus string

const (
	InstallationReportPending  InstallationReportStatus = "pending"
	InstallationReportVerified InstallationReportStatus = "verified"
	InstallationReportRejected InstallationReportStatus = "rejected"
)

func (r InstallationReportStatus) Outcome(riskScore, alertThreshold float64) InstallationReportStatus {
	if riskScore > alertThreshold {
		return InstallationReportRejected
	}
	return InstallationReportVerified
}
