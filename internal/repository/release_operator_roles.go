package repository

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

func (p *Postgres) ChangeReleaseOperatorRole(ctx context.Context, id string, role domain.ReleaseOperatorRole) error {
	if err := (domain.ReleaseOperator{ID: id, Name: "valid", Role: role}).Validate(); err != nil {
		return err
	}
	result, err := p.pool.Exec(ctx, `UPDATE release_operators SET role=$1 WHERE id=$2`, role, id)
	if err != nil {
		return fmt.Errorf("change release_operator role: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrNotFound
	}
	return nil
}

func (p *Postgres) ReleaseOperatorsForRole(ctx context.Context, role domain.ReleaseOperatorRole) ([]domain.ReleaseOperator, error) {
	rows, err := p.pool.Query(ctx, `SELECT id,name,role FROM release_operators WHERE role=$1 ORDER BY name`, role)
	if err != nil {
		return nil, fmt.Errorf("release_operators for role: %w", err)
	}
	defer rows.Close()
	items := make([]domain.ReleaseOperator, 0)
	for rows.Next() {
		var item domain.ReleaseOperator
		if err := rows.Scan(&item.ID, &item.Name, &item.Role); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
