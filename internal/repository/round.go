package repository

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

func (p *Postgres) StartRolloutWave(ctx context.Context, id string, version int64) error {
	return changeRolloutWave(ctx, p.pool, id, domain.RolloutWaveRunning, version)
}
func (p *Postgres) CompleteRolloutWave(ctx context.Context, id string, version int64) error {
	return changeRolloutWave(ctx, p.pool, id, domain.RolloutWaveCompleted, version)
}
func (p *Postgres) CancelRolloutWave(ctx context.Context, id string, version int64) error {
	return changeRolloutWave(ctx, p.pool, id, domain.RolloutWaveCancelled, version)
}
func (t *transaction) StartRolloutWave(ctx context.Context, id string, version int64) error {
	return changeRolloutWave(ctx, t.tx, id, domain.RolloutWaveRunning, version)
}
func (t *transaction) CompleteRolloutWave(ctx context.Context, id string, version int64) error {
	return changeRolloutWave(ctx, t.tx, id, domain.RolloutWaveCompleted, version)
}
func (t *transaction) CancelRolloutWave(ctx context.Context, id string, version int64) error {
	return changeRolloutWave(ctx, t.tx, id, domain.RolloutWaveCancelled, version)
}

func changeRolloutWave(ctx context.Context, q sqler, id string, status domain.RolloutWaveStatus, version int64) error {
	if status == domain.RolloutWaveCancelled {
		var changed int
		err := q.QueryRow(ctx, `WITH changed AS (
			UPDATE rollout_waves SET status='cancelled',version=version+1,completed_at=now()
			WHERE id=$1 AND version=$2 AND status IN ('queued','running')
			AND NOT EXISTS (SELECT 1 FROM installation_reports WHERE rollout_wave_id=$1)
			RETURNING id
		), restored AS (
			UPDATE deployment_jobs s SET status='accepted',version=version+1
			FROM rollout_wave_items bs, changed
			WHERE bs.rollout_wave_id=changed.id AND s.id=bs.deployment_job_id AND s.status='in_progress'
			RETURNING s.id
		)
		SELECT count(*) FROM changed`, id, version).Scan(&changed)
		if err != nil {
			return fmt.Errorf("cancel rollout_wave: %w", err)
		}
		if changed != 1 {
			return domain.ErrConflict
		}
		return nil
	}
	allowedFrom := []string{}
	switch status {
	case domain.RolloutWaveRunning:
		allowedFrom = []string{string(domain.RolloutWaveQueued)}
	case domain.RolloutWaveCompleted:
		allowedFrom = []string{string(domain.RolloutWaveRunning)}
	default:
		return domain.ErrInvalidTransition
	}
	result, err := q.Exec(ctx, `UPDATE rollout_waves SET status=$1,version=version+1,started_at=CASE WHEN $1='running' THEN now() ELSE started_at END,completed_at=CASE WHEN $1='completed' THEN now() ELSE completed_at END
		WHERE id=$2 AND version=$3 AND status=ANY($4)
		AND ($1 <> 'completed' OR NOT EXISTS (
			SELECT 1 FROM rollout_wave_items bs
			LEFT JOIN installation_reports r ON r.rollout_wave_id=bs.rollout_wave_id AND r.deployment_job_id=bs.deployment_job_id
			WHERE bs.rollout_wave_id=$2 AND (r.id IS NULL OR r.status='pending')
		))`, status, id, version, allowedFrom)
	if err != nil {
		return fmt.Errorf("change rollout_wave: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}
