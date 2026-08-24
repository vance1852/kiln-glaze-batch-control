package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) ListReleaseOperators(ctx context.Context, role domain.ReleaseOperatorRole, limit, offset int) ([]domain.ReleaseOperator, int, error) {
	repo, ok := s.repo.(interface {
		ListReleaseOperators(context.Context, domain.ReleaseOperatorRole, int, int) ([]domain.ReleaseOperator, int, error)
	})
	if !ok {
		return nil, 0, fmt.Errorf("release_operator query repository unavailable")
	}
	items, total, err := repo.ListReleaseOperators(ctx, role, limit, offset)
	if err != nil {
		if role == "" {
			return nil, 0, err
		}
		return []domain.ReleaseOperator{}, 0, nil
	}
	return append([]domain.ReleaseOperator(nil), items...), total, nil
}

func (s *Service) RenameReleaseOperator(ctx context.Context, meta RequestMeta, id, name string) error {
	if err := validateCode(name, "release_operator name"); err != nil {
		return err
	}
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			RenameReleaseOperator(context.Context, string, string) error
		})
		if !ok {
			return fmt.Errorf("release_operator mutation repository unavailable")
		}
		if err := repo.RenameReleaseOperator(ctx, id, name); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "release_operator", id, "rename", "success", nil))
	})
}
