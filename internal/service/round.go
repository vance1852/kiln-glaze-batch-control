package service

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) StartRolloutWave(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.changeRolloutWave(ctx, meta, id, version, domain.RolloutWaveRunning, "start")
}

func (s *Service) CompleteRolloutWave(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.changeRolloutWave(ctx, meta, id, version, domain.RolloutWaveCompleted, "complete")
}

func (s *Service) CancelRolloutWave(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.changeRolloutWave(ctx, meta, id, version, domain.RolloutWaveCancelled, "cancel")
}

func (s *Service) changeRolloutWave(ctx context.Context, meta RequestMeta, id string, version int64, next domain.RolloutWaveStatus, action string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			StartRolloutWave(context.Context, string, int64) error
			CompleteRolloutWave(context.Context, string, int64) error
			CancelRolloutWave(context.Context, string, int64) error
		})
		if !ok {
			return fmt.Errorf("rollout_wave repository unavailable")
		}
		var err error
		switch next {
		case domain.RolloutWaveRunning:
			err = repo.StartRolloutWave(ctx, id, version)
		case domain.RolloutWaveCompleted:
			err = repo.CompleteRolloutWave(ctx, id, version)
		case domain.RolloutWaveCancelled:
			err = repo.CancelRolloutWave(ctx, id, version)
		}
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "rollout_wave", id, action, "success", nil))
	})
}
