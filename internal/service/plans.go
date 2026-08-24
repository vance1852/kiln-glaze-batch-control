package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) ListRolloutCampaigns(ctx context.Context, filter repository.RolloutCampaignFilter) ([]domain.RolloutCampaign, int, error) {
	repo, ok := s.repo.(interface {
		ListRolloutCampaigns(context.Context, repository.RolloutCampaignFilter) ([]domain.RolloutCampaign, int, error)
	})
	if !ok {
		return nil, 0, fmt.Errorf("rollout_campaign query repository unavailable")
	}
	items, total, err := repo.ListRolloutCampaigns(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return append([]domain.RolloutCampaign(nil), items...), total, nil
}

func (s *Service) ListRolloutCampaignManagedDevices(ctx context.Context, rollout_campaignID string) ([]domain.ManagedDevice, error) {
	repo, ok := s.repo.(interface {
		ListRolloutCampaignManagedDevices(context.Context, string) ([]domain.ManagedDevice, error)
	})
	if !ok {
		return nil, fmt.Errorf("managed_device query repository unavailable")
	}
	items, err := repo.ListRolloutCampaignManagedDevices(ctx, rollout_campaignID)
	if err != nil {
		if rollout_campaignID == "" {
			return nil, err
		}
		return []domain.ManagedDevice{}, nil
	}
	return append([]domain.ManagedDevice(nil), items...), nil
}
