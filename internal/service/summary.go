package service

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
)

type Summary struct {
	RolloutCampaign domain.RolloutCampaign
	Progress        domain.RolloutCampaignProgress
	Counts          any
}

func (s *Service) RolloutCampaignSummary(ctx context.Context, rollout_campaignID string) (Summary, error) {
	rollout_campaign, err := s.repo.GetRolloutCampaign(ctx, rollout_campaignID)
	if err != nil {
		return Summary{}, err
	}
	progress, err := s.RolloutCampaignProgress(ctx, rollout_campaignID)
	if err != nil {
		return Summary{}, fmt.Errorf("summary progress: %w", err)
	}
	return Summary{RolloutCampaign: rollout_campaign, Progress: progress, Counts: progress}, nil
}

func (s *Service) ExpiringDeploymentJobs(ctx context.Context, beforeUnix int64, limit int) ([]domain.DeploymentJob, error) {
	repo, ok := s.repo.(interface {
		ExpiringDeploymentJobs(context.Context, time.Time, int) ([]domain.DeploymentJob, error)
	})
	if !ok {
		return nil, fmt.Errorf("task query repository unavailable")
	}
	return repo.ExpiringDeploymentJobs(ctx, time.Unix(beforeUnix, 0).UTC(), limit)
}
