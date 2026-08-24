package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

func (s *Service) Authorize(ctx context.Context, release_operatorID, action string) error {
	release_operator, err := s.LoadReleaseOperator(ctx, release_operatorID)
	if err != nil {
		return err
	}
	if !release_operator.Can(action) {
		return fmt.Errorf("release_operator cannot %s: %w", action, domain.ErrConflict)
	}
	return nil
}

func RequireSupervisor(ctx context.Context, s *Service, release_operatorID string) error {
	return s.Authorize(ctx, release_operatorID, "close_rollout_campaign")
}

func RequireReviewer(ctx context.Context, s *Service, release_operatorID string) error {
	return s.Authorize(ctx, release_operatorID, "review_installation_report")
}
