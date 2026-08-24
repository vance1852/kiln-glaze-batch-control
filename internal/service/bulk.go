package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) CreateDeploymentJobsBulk(ctx context.Context, meta RequestMeta, requests []domain.DeploymentJobRequest) ([]domain.BulkItemResult, error) {
	now := s.now()
	if len(requests) == 0 {
		return nil, fmt.Errorf("at least one task is required: %w", domain.ErrConflict)
	}
	if err := domain.ValidateBulkRequests(requests, now); err != nil {
		return nil, err
	}
	inputs := make([]repository.DeploymentJobInput, len(requests))
	for i, request := range requests {
		inputs[i] = repository.DeploymentJobInput{RolloutCampaignID: request.RolloutCampaignID, ManagedDeviceID: request.ManagedDeviceID, TaskCode: request.TaskCode, ExpiresAt: request.ExpiresAt}
	}
	var deployment_jobs []domain.DeploymentJob
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		for _, input := range inputs {
			placement, ok := tx.(interface {
				ValidateRolloutCampaignManagedDevice(context.Context, string, string) error
			})
			if !ok {
				return fmt.Errorf("task placement repository unavailable")
			}
			if err := placement.ValidateRolloutCampaignManagedDevice(ctx, input.RolloutCampaignID, input.ManagedDeviceID); err != nil {
				return err
			}
			task, err := tx.CreateDeploymentJob(ctx, input)
			if err != nil {
				return err
			}
			deployment_jobs = append(deployment_jobs, task)
		}
		return tx.WriteAudit(ctx, audit(meta, "task_rollout_wave", requests[0].RolloutCampaignID, "create_bulk", "success", nil))
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.BulkItemResult, len(deployment_jobs))
	for i, task := range deployment_jobs {
		result[i] = domain.BulkItemResult{Index: i, TaskCode: task.TaskCode, DeploymentJobID: task.ID}
	}
	return result, nil
}

func (s *Service) ValidateBulkForManagedDevice(requests []domain.DeploymentJobRequest, managed_deviceID string) error {
	if err := domain.ValidateBulkRequests(requests, s.now()); err != nil {
		return err
	}
	for _, request := range requests {
		if request.ManagedDeviceID != managed_deviceID {
			return fmt.Errorf("task managed_device differs from rollout_wave managed_device: %w", domain.ErrConflict)
		}
	}
	return nil
}
