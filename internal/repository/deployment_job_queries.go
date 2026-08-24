package repository

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
)

type DeploymentJobCounts struct {
	Queued     int
	Completed  int
	Transit    int
	Accepted   int
	InProgress int
	Verified   int
	Rejected   int
	Archived   int
}

func (p *Postgres) DeploymentJobCounts(ctx context.Context, rollout_campaignID string) (DeploymentJobCounts, error) {
	var counts DeploymentJobCounts
	err := p.pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status='queued'),count(*) FILTER (WHERE status='completed'),count(*) FILTER (WHERE status='activation_pending'),count(*) FILTER (WHERE status='accepted'),count(*) FILTER (WHERE status='in_progress'),count(*) FILTER (WHERE status='verified'),count(*) FILTER (WHERE status='rejected'),count(*) FILTER (WHERE status='archived') FROM deployment_jobs WHERE rollout_campaign_id=$1`, rollout_campaignID).Scan(&counts.Queued, &counts.Completed, &counts.Transit, &counts.Accepted, &counts.InProgress, &counts.Verified, &counts.Rejected, &counts.Archived)
	if err != nil {
		return DeploymentJobCounts{}, fmt.Errorf("task counts: %w", err)
	}
	return counts, nil
}

func (p *Postgres) ExpiringDeploymentJobs(ctx context.Context, before time.Time, limit int) ([]domain.DeploymentJob, error) {
	return p.expiringDeploymentJobs(ctx, "", before, limit)
}

func (p *Postgres) ExpiringDeploymentJobsForRolloutCampaign(ctx context.Context, rollout_campaignID string, before time.Time, limit int) ([]domain.DeploymentJob, error) {
	return p.expiringDeploymentJobs(ctx, rollout_campaignID, before, limit)
}

func (p *Postgres) expiringDeploymentJobs(ctx context.Context, rollout_campaignID string, before time.Time, limit int) ([]domain.DeploymentJob, error) {
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	query := `SELECT id,rollout_campaign_id,managed_device_id,task_code,status,completed_at,accepted_at,expires_at,version FROM deployment_jobs WHERE status NOT IN ('verified','archived') AND expires_at <= $1`
	args := []any{before, limit}
	if rollout_campaignID != "" {
		query += " AND rollout_campaign_id=$3"
		args = append(args, rollout_campaignID)
	}
	query += " ORDER BY expires_at LIMIT $2"
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("expiring deployment_jobs: %w", err)
	}
	defer rows.Close()
	items := make([]domain.DeploymentJob, 0)
	for rows.Next() {
		var task domain.DeploymentJob
		if err := rows.Scan(&task.ID, &task.RolloutCampaignID, &task.ManagedDeviceID, &task.TaskCode, &task.Status, &task.CompletedAt, &task.AcceptedAt, &task.ExpiresAt, &task.Version); err != nil {
			return nil, fmt.Errorf("scan expiring task: %w", err)
		}
		items = append(items, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read expiring deployment_jobs: %w", err)
	}
	return items, nil
}
