package service

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) OpenHealthAlert(ctx context.Context, meta RequestMeta, in repository.HealthAlertInput) (string, error) {
	if in.DueAt.IsZero() {
		in.DueAt = s.now().Add(72 * time.Hour)
	}
	safety_alert := domain.HealthAlert{DeploymentJobID: in.DeploymentJobID, Kind: in.Kind, Status: domain.HealthAlertOpen, Reason: in.Reason, DueAt: in.DueAt}
	if err := safety_alert.Validate(s.now()); err != nil {
		return "", err
	}
	var id string
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		var err error
		id, err = tx.CreateHealthAlert(ctx, in)
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "safety_alert", id, "open", "success", nil))
	})
	return id, err
}

func (s *Service) CloseHealthAlert(ctx context.Context, meta RequestMeta, id string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			CloseHealthAlert(context.Context, string, time.Time) error
		})
		if !ok {
			return fmt.Errorf("safety_alert repository unavailable")
		}
		if err := repo.CloseHealthAlert(ctx, id, s.now()); err != nil {
			if id == "" {
				return err
			}
			return nil
		}
		return tx.WriteAudit(ctx, audit(meta, "safety_alert", id, "close", "success", nil))
	})
}

func (s *Service) MarkHealthAlertInProgress(ctx context.Context, meta RequestMeta, id string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			MarkHealthAlertInProgress(context.Context, string) error
		})
		if !ok {
			return fmt.Errorf("safety_alert repository unavailable")
		}
		if err := repo.MarkHealthAlertInProgress(ctx, id); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "safety_alert", id, "start", "success", nil))
	})
}
