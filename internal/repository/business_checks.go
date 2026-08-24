package repository

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) ValidateRolloutCampaignManagedDevice(ctx context.Context, rollout_campaignID, managed_deviceID string) error {
	return validateRolloutCampaignManagedDevice(ctx, p.pool, rollout_campaignID, managed_deviceID)
}

func (t *transaction) ValidateRolloutCampaignManagedDevice(ctx context.Context, rollout_campaignID, managed_deviceID string) error {
	return validateRolloutCampaignManagedDevice(ctx, t.tx, rollout_campaignID, managed_deviceID)
}

func validateRolloutCampaignManagedDevice(ctx context.Context, q sqler, rollout_campaignID, managed_deviceID string) error {
	var exists bool
	if err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM managed_devices WHERE id=$1 AND rollout_campaign_id=$2)`, managed_deviceID, rollout_campaignID).Scan(&exists); err != nil {
		return fmt.Errorf("validate rollout_campaign managed_device: %w", err)
	}
	if !exists {
		return fmt.Errorf("managed_device does not belong to rollout_campaign: %w", domain.ErrConflict)
	}
	return nil
}

func (p *Postgres) ValidateInstallationReportTarget(ctx context.Context, taskID, rollout_waveID string) error {
	return validateInstallationReportTarget(ctx, p.pool, taskID, rollout_waveID)
}

func (t *transaction) ValidateInstallationReportTarget(ctx context.Context, taskID, rollout_waveID string) error {
	return validateInstallationReportTarget(ctx, t.tx, taskID, rollout_waveID)
}

func validateInstallationReportTarget(ctx context.Context, q sqler, taskID, rollout_waveID string) error {
	var taskStatus, rollout_waveStatus string
	err := q.QueryRow(ctx, `SELECT s.status,b.status FROM rollout_wave_items bs JOIN deployment_jobs s ON s.id=bs.deployment_job_id JOIN rollout_waves b ON b.id=bs.rollout_wave_id WHERE bs.deployment_job_id=$1 AND bs.rollout_wave_id=$2`, taskID, rollout_waveID).Scan(&taskStatus, &rollout_waveStatus)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("task is not attached to rollout_wave: %w", domain.ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("validate installation_report target: %w", err)
	}
	if taskStatus != string(domain.DeploymentJobInProgress) || rollout_waveStatus != string(domain.RolloutWaveRunning) {
		return fmt.Errorf("task and managed_device round are not ready for an installation_report: %w", domain.ErrInvalidTransition)
	}
	return nil
}

func (p *Postgres) InstallationReportTaskID(ctx context.Context, installation_reportID string) (string, error) {
	return installation_reportDeploymentJobID(ctx, p.pool, installation_reportID)
}

func (t *transaction) InstallationReportTaskID(ctx context.Context, installation_reportID string) (string, error) {
	return installation_reportDeploymentJobID(ctx, t.tx, installation_reportID)
}

func installation_reportDeploymentJobID(ctx context.Context, q sqler, installation_reportID string) (string, error) {
	var taskID string
	if err := q.QueryRow(ctx, `SELECT deployment_job_id FROM installation_reports WHERE id=$1`, installation_reportID).Scan(&taskID); err == pgx.ErrNoRows {
		return "", domain.ErrNotFound
	} else if err != nil {
		return "", fmt.Errorf("get installation_report task: %w", err)
	}
	return taskID, nil
}
