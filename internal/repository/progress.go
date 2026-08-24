package repository

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

func (p *Postgres) RolloutCampaignProgress(ctx context.Context, rollout_campaignID string) (domain.RolloutCampaignProgress, error) {
	return rollout_campaignProgress(ctx, p.pool, rollout_campaignID)
}

func rollout_campaignProgress(ctx context.Context, q sqler, rollout_campaignID string) (domain.RolloutCampaignProgress, error) {
	var progress domain.RolloutCampaignProgress
	progress.RolloutCampaignID = rollout_campaignID
	err := q.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM managed_devices WHERE rollout_campaign_id=$1),
		COALESCE((SELECT sum(required_successes) FROM managed_devices WHERE rollout_campaign_id=$1),0),
		count(*) FILTER (WHERE status='completed'),
		count(*) FILTER (WHERE status='accepted'),
		count(*) FILTER (WHERE status='in_progress'),
		count(*) FILTER (WHERE status='verified'),
		count(*) FILTER (WHERE status='rejected'),
		count(*) FILTER (WHERE status='archived')
		FROM deployment_jobs WHERE rollout_campaign_id=$1`, rollout_campaignID).Scan(&progress.ManagedDevices, &progress.Required, &progress.Completed, &progress.Accepted, &progress.InProgress, &progress.Verified, &progress.Rejected, &progress.Archived)
	if err != nil {
		return domain.RolloutCampaignProgress{}, fmt.Errorf("rollout_campaign progress: %w", err)
	}
	return progress, nil
}
