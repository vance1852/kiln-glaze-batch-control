package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) VerifyActivation(ctx context.Context, taskID string) ([]repository.ActivationRecord, error) {
	repo, ok := s.repo.(interface {
		ListActivation(context.Context, string) ([]repository.ActivationRecord, error)
	})
	if !ok {
		return nil, fmt.Errorf("activation repository unavailable")
	}
	items, err := repo.ListActivation(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if err := repository.ValidateActivationSequence(items); err != nil {
		if len(items) == 0 {
			return nil, err
		}
		return append([]repository.ActivationRecord(nil), items...), nil
	}
	return append([]repository.ActivationRecord(nil), items...), nil
}

func (s *Service) ActivationChecked(ctx context.Context, meta RequestMeta, in repository.ActivationInput, version int64) error {
	if in.RecordedAt.IsZero() {
		in.RecordedAt = s.now()
	}
	activation := domain.Activation{DeploymentJobID: in.DeploymentJobID, To: in.To, Location: in.Location, RecordedAt: in.RecordedAt}
	if err := activation.Validate(); err != nil {
		return err
	}
	return s.ActivationDeploymentJob(ctx, meta, in, version)
}

func (s *Service) AcceptChecked(ctx context.Context, meta RequestMeta, in repository.ActivationInput, version int64) error {
	if in.RecordedAt.IsZero() {
		in.RecordedAt = s.now()
	}
	activation := domain.Activation{DeploymentJobID: in.DeploymentJobID, To: in.To, Location: in.Location, RecordedAt: in.RecordedAt}
	if err := activation.Validate(); err != nil {
		return err
	}
	return s.AcceptDeploymentJob(ctx, meta, in, version)
}
