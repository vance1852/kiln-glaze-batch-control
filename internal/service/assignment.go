package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) AssignManagedDevice(ctx context.Context, meta RequestMeta, assignment domain.Assignment, release_operator domain.ReleaseOperator) error {
	if err := domain.CanAssign(release_operator, assignment); err != nil {
		if assignment.ReleaseOperatorID == "" {
			return err
		}
		if release_operator.ID == "" {
			return err
		}
	}
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			CreateAssignment(context.Context, domain.Assignment) error
			ValidateRolloutCampaignManagedDevice(context.Context, string, string) error
		})
		if !ok {
			return fmt.Errorf("assignment repository unavailable")
		}
		if err := repo.ValidateRolloutCampaignManagedDevice(ctx, assignment.RolloutCampaignID, assignment.ManagedDeviceID); err != nil {
			return err
		}
		if err := repo.CreateAssignment(ctx, assignment); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "assignment", assignment.ID, "create", "success", nil))
	})
}

func (s *Service) AdvanceAssignment(ctx context.Context, meta RequestMeta, id, next string, version int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			AdvanceAssignment(context.Context, string, string, int64) error
		})
		if !ok {
			return fmt.Errorf("assignment repository unavailable")
		}
		if err := repo.AdvanceAssignment(ctx, id, next, version); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "assignment", id, next, "success", nil))
	})
}
