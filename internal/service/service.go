package service

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
	"github.com/google/uuid"
)

type Clock func() time.Time

type Service struct {
	repo repository.Repository
	now  Clock
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) WithClock(clock Clock) *Service {
	if clock != nil {
		s.now = clock
	}
	return s
}

type RequestMeta struct {
	RequestID         string
	ReleaseOperatorID *string
}

type CreateRolloutCampaignRequest struct {
	Code           string
	Name           string
	Timezone       string
	StartsAt       time.Time
	EndsAt         time.Time
	CreatedBy      string
	ManagedDevices []repository.ManagedDeviceInput
}

type CreateRolloutCampaignResponse struct {
	RolloutCampaign  domain.RolloutCampaign `json:"rollout_campaign"`
	ManagedDeviceIDs []string               `json:"managed_device_ids"`
}

func (s *Service) CreateRolloutCampaign(ctx context.Context, meta RequestMeta, in CreateRolloutCampaignRequest) (CreateRolloutCampaignResponse, error) {
	if in.Code == "" || in.Name == "" || in.CreatedBy == "" || len(in.ManagedDevices) == 0 {
		return CreateRolloutCampaignResponse{}, fmt.Errorf("code, name, creator and managed_devices are required: %w", domain.ErrConflict)
	}
	if err := domain.ValidateBusinessCode(in.Code); err != nil {
		return CreateRolloutCampaignResponse{}, err
	}
	if err := validateCode(in.Name, "rollout_campaign name"); err != nil {
		return CreateRolloutCampaignResponse{}, err
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return CreateRolloutCampaignResponse{}, fmt.Errorf("invalid rollout_campaign timezone: %w", domain.ErrConflict)
	}
	seenManagedDevices := make(map[string]struct{}, len(in.ManagedDevices))
	for _, managed_device := range in.ManagedDevices {
		if err := (domain.ManagedDevice{Code: managed_device.Code, RolloutLane: managed_device.RolloutLane, RequiredSuccesses: managed_device.RequiredSuccesses}).Validate(); err != nil {
			return CreateRolloutCampaignResponse{}, err
		}
		if err := domain.ValidateBusinessCode(managed_device.Code); err != nil {
			return CreateRolloutCampaignResponse{}, err
		}
		if _, exists := seenManagedDevices[managed_device.Code]; exists {
			return CreateRolloutCampaignResponse{}, fmt.Errorf("duplicate managed_device code %s: %w", managed_device.Code, domain.ErrConflict)
		}
		seenManagedDevices[managed_device.Code] = struct{}{}
	}
	rollout_campaign := domain.RolloutCampaign{ID: uuid.NewString(), Code: in.Code, Name: in.Name, Status: domain.RolloutCampaignDraft, Timezone: in.Timezone, StartsAt: in.StartsAt, EndsAt: in.EndsAt, Version: 1, CreatedBy: in.CreatedBy}
	if err := rollout_campaign.ValidateWindow(s.now()); err != nil {
		return CreateRolloutCampaignResponse{}, err
	}
	response := CreateRolloutCampaignResponse{RolloutCampaign: rollout_campaign, ManagedDeviceIDs: make([]string, 0, len(in.ManagedDevices))}
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		if err := tx.CreateRolloutCampaign(ctx, &rollout_campaign); err != nil {
			return err
		}
		for _, managed_device := range in.ManagedDevices {
			managed_device.RolloutCampaignID = rollout_campaign.ID
			id, err := tx.CreateManagedDevice(ctx, managed_device)
			if err != nil {
				return err
			}
			response.ManagedDeviceIDs = append(response.ManagedDeviceIDs, id)
		}
		return tx.WriteAudit(ctx, audit(meta, "managed_device_rollout_campaign", rollout_campaign.ID, "create", "success", nil))
	})
	if err != nil {
		return CreateRolloutCampaignResponse{}, err
	}
	return response, nil
}

func (s *Service) ScheduleRolloutCampaign(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.advanceRolloutCampaign(ctx, meta, id, domain.RolloutCampaignScheduled, version, "schedule")
}

func (s *Service) ActivateRolloutCampaign(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.advanceRolloutCampaign(ctx, meta, id, domain.RolloutCampaignActive, version, "activate")
}

func (s *Service) CloseRolloutCampaign(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.advanceRolloutCampaign(ctx, meta, id, domain.RolloutCampaignClosed, version, "close")
}

func (s *Service) advanceRolloutCampaign(ctx context.Context, meta RequestMeta, id string, next domain.RolloutCampaignStatus, version int64, action string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		rollout_campaign, err := tx.GetRolloutCampaign(ctx, id)
		if err != nil {
			return err
		}
		if !rollout_campaign.Status.CanMoveTo(next) {
			return fmt.Errorf("rollout_campaign %s: %w", id, domain.ErrInvalidTransition)
		}
		if err := tx.AdvanceRolloutCampaign(ctx, id, next, version); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "managed_device_rollout_campaign", id, action, "success", nil))
	})
}

func (s *Service) CreateDeploymentJob(ctx context.Context, meta RequestMeta, in repository.DeploymentJobInput) (domain.DeploymentJob, error) {
	request := domain.DeploymentJobRequest{RolloutCampaignID: in.RolloutCampaignID, ManagedDeviceID: in.ManagedDeviceID, TaskCode: in.TaskCode, ExpiresAt: in.ExpiresAt}
	if err := request.Validate(s.now()); err != nil {
		return domain.DeploymentJob{}, err
	}
	if err := domain.ValidateBusinessCode(in.TaskCode); err != nil {
		return domain.DeploymentJob{}, err
	}
	var task domain.DeploymentJob
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		placement, ok := tx.(interface {
			ValidateRolloutCampaignManagedDevice(context.Context, string, string) error
		})
		if !ok {
			return fmt.Errorf("task placement repository unavailable")
		}
		if err := placement.ValidateRolloutCampaignManagedDevice(ctx, in.RolloutCampaignID, in.ManagedDeviceID); err != nil {
			return err
		}
		var err error
		task, err = tx.CreateDeploymentJob(ctx, in)
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "managed_device_task", task.ID, "create", "success", nil))
	})
	return task, err
}

func (s *Service) CompleteDeploymentJob(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.moveDeploymentJob(ctx, meta, id, version, domain.DeploymentJobCompleted, "complete")
}

func (s *Service) ActivationDeploymentJob(ctx context.Context, meta RequestMeta, in repository.ActivationInput, version int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		if err := tx.MoveDeploymentJob(ctx, in.DeploymentJobID, domain.DeploymentJobActivationPending, version, in.RecordedAt); err != nil {
			return err
		}
		if err := tx.RecordActivation(ctx, in); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "managed_device_task", in.DeploymentJobID, "activation", "success", nil))
	})
}

func (s *Service) AcceptDeploymentJob(ctx context.Context, meta RequestMeta, in repository.ActivationInput, version int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		if err := tx.MoveDeploymentJob(ctx, in.DeploymentJobID, domain.DeploymentJobAccepted, version, in.RecordedAt); err != nil {
			return err
		}
		if err := tx.RecordActivation(ctx, in); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "managed_device_task", in.DeploymentJobID, "accept", "success", nil))
	})
}

func (s *Service) CreateRolloutWave(ctx context.Context, meta RequestMeta, in repository.RolloutWaveInput, taskIDs []string) (string, error) {
	if len(taskIDs) == 0 || len(taskIDs) > in.Capacity {
		return "", domain.ErrCapacityExceeded
	}
	if err := (domain.RolloutWave{Code: in.Code, Method: in.Method, Capacity: in.Capacity, DeploymentJobs: taskIDs}).Validate(); err != nil {
		return "", err
	}
	if err := domain.ValidateBusinessCode(in.Code); err != nil {
		return "", err
	}
	if err := validateIDs(taskIDs); err != nil {
		return "", err
	}
	var id string
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		var err error
		id, err = tx.CreateRolloutWave(ctx, in)
		if err != nil {
			return err
		}
		if err := tx.AttachDeploymentJobs(ctx, id, append([]string(nil), taskIDs...)); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "rollout_wave", id, "create", "success", nil))
	})
	return id, err
}

func (s *Service) SubmitInstallationReport(ctx context.Context, meta RequestMeta, in repository.InstallationReportInput) (string, error) {
	if in.DeploymentJobID == "" || in.RolloutWaveID == "" || in.RecorderID == "" || in.ObservedAt.IsZero() {
		return "", fmt.Errorf("installation_report identifiers and observed_at are required: %w", domain.ErrConflict)
	}
	if err := domain.ValidateInstallationReport(in.RiskScore, in.AlertThreshold, in.Scale); err != nil {
		return "", err
	}
	var id string
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		target, ok := tx.(interface {
			ValidateInstallationReportTarget(context.Context, string, string) error
			GetReleaseOperator(context.Context, string) (domain.ReleaseOperator, error)
		})
		if !ok {
			return fmt.Errorf("installation_report target repository unavailable")
		}
		if err := target.ValidateInstallationReportTarget(ctx, in.DeploymentJobID, in.RolloutWaveID); err != nil {
			return err
		}
		recorder, err := target.GetReleaseOperator(ctx, in.RecorderID)
		if err != nil {
			return err
		}
		if !recorder.CanRecordInstallationReport() {
			return fmt.Errorf("release_operator role %s cannot record installation_reports: %w", recorder.Role, domain.ErrConflict)
		}
		id, err = tx.CreateInstallationReport(ctx, in)
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "installation_report", id, "record", "success", nil))
	})
	return id, err
}

func (s *Service) ReviewInstallationReport(ctx context.Context, meta RequestMeta, installation_reportID, taskID string, accepted bool, installation_reportVersion, taskVersion int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		installation_report, ok := tx.(interface {
			InstallationReportTaskID(context.Context, string) (string, error)
		})
		if !ok {
			return fmt.Errorf("installation_report query repository unavailable")
		}
		actualDeploymentJobID, err := installation_report.InstallationReportTaskID(ctx, installation_reportID)
		if err != nil {
			return err
		}
		if actualDeploymentJobID != taskID {
			return fmt.Errorf("installation_report and task do not match: %w", domain.ErrConflict)
		}
		if err := tx.ReviewInstallationReportRecord(ctx, installation_reportID, accepted, installation_reportVersion, s.now()); err != nil {
			return err
		}
		next := domain.DeploymentJobVerified
		if !accepted {
			next = domain.DeploymentJobRejected
		}
		if err := tx.MoveDeploymentJob(ctx, taskID, next, taskVersion, s.now()); err != nil {
			return err
		}
		if !accepted {
			safety_alertID, err := tx.CreateHealthAlert(ctx, repository.HealthAlertInput{DeploymentJobID: taskID, Kind: "reassess", Reason: "risk score exceeded the alert threshold", DueAt: s.now().Add(72 * time.Hour)})
			if err != nil {
				return err
			}
			if err := tx.WriteAudit(ctx, audit(meta, "safety_alert", safety_alertID, "open", "success", nil)); err != nil {
				return err
			}
		}
		return tx.WriteAudit(ctx, audit(meta, "installation_report", installation_reportID, "review", "success", nil))
	})
}

func (s *Service) ArchiveDeploymentJob(ctx context.Context, meta RequestMeta, taskID string, version int64) error {
	return s.moveDeploymentJob(ctx, meta, taskID, version, domain.DeploymentJobArchived, "archive")
}

func (s *Service) ListDeploymentJobs(ctx context.Context, offset, limit int, rollout_campaignID string, status domain.DeploymentJobStatus) (repository.Page, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListDeploymentJobs(ctx, offset, limit, rollout_campaignID, status)
}

func (s *Service) DueHealthAlerts(ctx context.Context, before time.Time, limit int) ([]repository.HealthAlertInput, error) {
	return s.repo.DueHealthAlerts(ctx, before, limit)
}

func (s *Service) moveDeploymentJob(ctx context.Context, meta RequestMeta, id string, version int64, next domain.DeploymentJobStatus, action string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		task, err := tx.GetDeploymentJob(ctx, id)
		if err != nil {
			return err
		}
		if next == domain.DeploymentJobCompleted {
			rollout_campaign, err := tx.GetRolloutCampaign(ctx, task.RolloutCampaignID)
			if err != nil {
				return err
			}
			if !rollout_campaign.CanExecuteAt(s.now()) {
				return fmt.Errorf("rollout_campaign is not active for task execution: %w", domain.ErrInvalidTransition)
			}
		}
		updated, err := task.Move(next, s.now())
		if err != nil {
			return err
		}
		if err := tx.MoveDeploymentJob(ctx, id, updated.Status, version, s.now()); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "managed_device_task", id, action, "success", nil))
	})
}

func audit(meta RequestMeta, objectType, objectID, action, outcome string, detail []byte) repository.AuditInput {
	return repository.AuditInput{RequestID: meta.RequestID, ReleaseOperatorID: meta.ReleaseOperatorID, ObjectType: objectType, ObjectID: objectID, Action: action, Outcome: outcome, Detail: detail}
}
