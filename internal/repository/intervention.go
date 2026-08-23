package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CloseHealthAlert(ctx context.Context, id string, now time.Time) error {
	return closeHealthAlert(ctx, p.pool, id, now)
}
func (t *transaction) CloseHealthAlert(ctx context.Context, id string, now time.Time) error {
	return closeHealthAlert(ctx, t.tx, id, now)
}

func (p *Postgres) GetHealthAlert(ctx context.Context, id string) (domain.HealthAlert, error) {
	var d domain.HealthAlert
	err := p.pool.QueryRow(ctx, `SELECT id,deployment_job_id,kind,status,reason,due_at,closed_at FROM health_alerts WHERE id=$1`, id).Scan(&d.ID, &d.DeploymentJobID, &d.Kind, &d.Status, &d.Reason, &d.DueAt, &d.ClosedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.HealthAlert{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.HealthAlert{}, fmt.Errorf("get safety_alert: %w", err)
	}
	return d, nil
}

func closeHealthAlert(ctx context.Context, q sqler, id string, now time.Time) error {
	result, err := q.Exec(ctx, `UPDATE health_alerts SET status='closed',closed_at=$1 WHERE id=$2 AND status IN ('open','in_progress')`, now, id)
	if err != nil {
		return fmt.Errorf("close safety_alert: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (p *Postgres) MarkHealthAlertInProgress(ctx context.Context, id string) error {
	return markHealthAlertInProgress(ctx, p.pool, id)
}

func (t *transaction) MarkHealthAlertInProgress(ctx context.Context, id string) error {
	return markHealthAlertInProgress(ctx, t.tx, id)
}

func markHealthAlertInProgress(ctx context.Context, q sqler, id string) error {
	result, err := q.Exec(ctx, `UPDATE health_alerts SET status='in_progress' WHERE id=$1 AND status='open'`, id)
	if err != nil {
		return fmt.Errorf("mark safety_alert in progress: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}
