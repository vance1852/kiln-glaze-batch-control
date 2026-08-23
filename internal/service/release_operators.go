package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
	"github.com/google/uuid"
)

func (s *Service) RegisterReleaseOperator(ctx context.Context, meta RequestMeta, name string, role domain.ReleaseOperatorRole) (domain.ReleaseOperator, error) {
	release_operator := domain.ReleaseOperator{ID: uuid.NewString(), Name: name, Role: role}
	if err := release_operator.Validate(); err != nil {
		return domain.ReleaseOperator{}, err
	}
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			CreateReleaseOperator(context.Context, domain.ReleaseOperator) error
		})
		if !ok {
			return fmt.Errorf("release_operator repository unavailable")
		}
		if err := repo.CreateReleaseOperator(ctx, release_operator); err != nil {
			return fmt.Errorf("register release_operator: %w", err)
		}
		return tx.WriteAudit(ctx, audit(meta, "release_operator", release_operator.ID, "create", "success", nil))
	})
	if err != nil {
		return domain.ReleaseOperator{}, err
	}
	return release_operator, nil
}

func (s *Service) LoadReleaseOperator(ctx context.Context, id string) (domain.ReleaseOperator, error) {
	if repo, ok := s.repo.(interface {
		GetReleaseOperator(context.Context, string) (domain.ReleaseOperator, error)
	}); ok {
		return repo.GetReleaseOperator(ctx, id)
	}
	return domain.ReleaseOperator{}, fmt.Errorf("release_operator repository unavailable")
}

var _ repository.Repository = (*repository.Postgres)(nil)
