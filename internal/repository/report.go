package repository

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
)

func (p *Postgres) OpenHealthAlertCount(ctx context.Context, rollout_campaignID string) (int, error) {
	var count int
	if err := p.pool.QueryRow(ctx, `SELECT count(*) FROM health_alerts d JOIN deployment_jobs s ON s.id=d.deployment_job_id WHERE s.rollout_campaign_id=$1 AND d.status IN ('open','in_progress')`, rollout_campaignID).Scan(&count); err != nil {
		return 0, fmt.Errorf("open safety_alert count: %w", err)
	}
	return count, nil
}

func (p *Postgres) ComplianceReport(ctx context.Context, rollout_campaignID string, now time.Time) (domain.ComplianceReport, error) {
	progress, err := p.RolloutCampaignProgress(ctx, rollout_campaignID)
	if err != nil {
		return domain.ComplianceReport{}, err
	}
	expiring, err := p.ExpiringDeploymentJobsForRolloutCampaign(ctx, rollout_campaignID, now.Add(48*time.Hour), 100)
	if err != nil {
		return domain.ComplianceReport{}, err
	}
	count, err := p.OpenHealthAlertCount(ctx, rollout_campaignID)
	if err != nil {
		return domain.ComplianceReport{}, err
	}
	return domain.ComplianceReport{RolloutCampaignID: rollout_campaignID, GeneratedAt: now.UTC(), Progress: progress, Expiring: expiring, OpenHealthAlerts: count}, nil
}
