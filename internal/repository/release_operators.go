package repository

import (
	"context"
	"errors"
	"fmt"

	"firmware-rollout-control/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CreateReleaseOperator(ctx context.Context, release_operator domain.ReleaseOperator) error {
	return createReleaseOperator(ctx, p.pool, release_operator)
}
func (t *transaction) CreateReleaseOperator(ctx context.Context, release_operator domain.ReleaseOperator) error {
	return createReleaseOperator(ctx, t.tx, release_operator)
}

func (p *Postgres) GetReleaseOperator(ctx context.Context, id string) (domain.ReleaseOperator, error) {
	return getReleaseOperator(ctx, p.pool, id)
}
func (t *transaction) GetReleaseOperator(ctx context.Context, id string) (domain.ReleaseOperator, error) {
	return getReleaseOperator(ctx, t.tx, id)
}

func createReleaseOperator(ctx context.Context, q sqler, release_operator domain.ReleaseOperator) error {
	if err := release_operator.Validate(); err != nil {
		return err
	}
	_, err := q.Exec(ctx, `INSERT INTO release_operators(id,name,role) VALUES ($1,$2,$3)`, release_operator.ID, release_operator.Name, release_operator.Role)
	return wrapWrite(err)
}

func getReleaseOperator(ctx context.Context, q sqler, id string) (domain.ReleaseOperator, error) {
	var release_operator domain.ReleaseOperator
	err := q.QueryRow(ctx, `SELECT id,name,role FROM release_operators WHERE id=$1`, id).Scan(&release_operator.ID, &release_operator.Name, &release_operator.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ReleaseOperator{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ReleaseOperator{}, fmt.Errorf("get release_operator: %w", err)
	}
	return release_operator, nil
}
