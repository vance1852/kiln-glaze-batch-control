package repository

import (
	"context"
	"time"

	"firmware-rollout-control/internal/domain"
)

type Page struct {
	Items  []domain.DeploymentJob `json:"items"`
	Total  int                    `json:"total"`
	Offset int                    `json:"offset"`
	Limit  int                    `json:"limit"`
}

type ManagedDeviceInput struct {
	RolloutCampaignID string `json:"rollout_campaign_id"`
	Code              string `json:"code"`
	RolloutLane       string `json:"rollout_lane"`
	RequiredSuccesses int    `json:"required_successes"`
}

type DeploymentJobInput struct {
	RolloutCampaignID string    `json:"rollout_campaign_id"`
	ManagedDeviceID   string    `json:"managed_device_id"`
	TaskCode          string    `json:"task_code"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type ActivationInput struct {
	DeploymentJobID string
	From            *string
	To              string
	Location        string
	RecordedAt      time.Time
	Note            string
}

type RolloutWaveInput struct {
	Code     string `json:"code"`
	Method   string `json:"method"`
	Capacity int    `json:"capacity"`
}

type InstallationReportInput struct {
	DeploymentJobID string    `json:"deployment_job_id"`
	RolloutWaveID   string    `json:"rollout_wave_id"`
	RecorderID      string    `json:"recorded_by"`
	RiskScore       float64   `json:"risk_score"`
	Scale           string    `json:"scale"`
	AlertThreshold  float64   `json:"alert_threshold"`
	ObservedAt      time.Time `json:"observed_at"`
}

type HealthAlertInput struct {
	DeploymentJobID string    `json:"deployment_job_id"`
	Kind            string    `json:"kind"`
	Reason          string    `json:"reason"`
	DueAt           time.Time `json:"due_at"`
}

type AuditInput struct {
	RequestID         string
	ReleaseOperatorID *string
	ObjectType        string
	ObjectID          string
	Action            string
	Outcome           string
	Detail            []byte
}

type Repository interface {
	InTx(context.Context, func(Repository) error) error
	CreateRolloutCampaign(context.Context, *domain.RolloutCampaign) error
	GetRolloutCampaign(context.Context, string) (domain.RolloutCampaign, error)
	AdvanceRolloutCampaign(context.Context, string, domain.RolloutCampaignStatus, int64) error
	CreateManagedDevice(context.Context, ManagedDeviceInput) (string, error)
	CreateDeploymentJob(context.Context, DeploymentJobInput) (domain.DeploymentJob, error)
	GetDeploymentJob(context.Context, string) (domain.DeploymentJob, error)
	MoveDeploymentJob(context.Context, string, domain.DeploymentJobStatus, int64, time.Time) error
	RecordActivation(context.Context, ActivationInput) error
	CreateRolloutWave(context.Context, RolloutWaveInput) (string, error)
	AttachDeploymentJobs(context.Context, string, []string) error
	CreateInstallationReport(context.Context, InstallationReportInput) (string, error)
	ReviewInstallationReportRecord(context.Context, string, bool, int64, time.Time) error
	CreateHealthAlert(context.Context, HealthAlertInput) (string, error)
	ListDeploymentJobs(context.Context, int, int, string, domain.DeploymentJobStatus) (Page, error)
	DueHealthAlerts(context.Context, time.Time, int) ([]HealthAlertInput, error)
	WriteAudit(context.Context, AuditInput) error
	Close() error
}
