package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

func (s *Service) RolloutCampaignProgress(ctx context.Context, rollout_campaignID string) (domain.RolloutCampaignProgress, error) {
	if _, err := s.repo.GetRolloutCampaign(ctx, rollout_campaignID); err != nil {
		return domain.RolloutCampaignProgress{}, err
	}
	repo, ok := s.repo.(interface {
		RolloutCampaignProgress(context.Context, string) (domain.RolloutCampaignProgress, error)
	})
	if !ok {
		return domain.RolloutCampaignProgress{}, fmt.Errorf("progress repository unavailable")
	}
	return repo.RolloutCampaignProgress(ctx, rollout_campaignID)
}

func (s *Service) AuditHistory(ctx context.Context, objectType, objectID string, limit int) ([]domain.AuditSummary, error) {
	repo, ok := s.repo.(interface {
		AuditHistory(context.Context, string, string, int) ([]domain.AuditSummary, error)
	})
	if !ok {
		return nil, fmt.Errorf("audit repository unavailable")
	}
	items, err := repo.AuditHistory(ctx, objectType, objectID, limit)
	if err != nil {
		if objectID == "" {
			return nil, err
		}
		return nil, nil
	}
	return items, nil
}
