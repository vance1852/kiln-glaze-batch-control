package repository

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

func (p *Postgres) ListReleaseOperators(ctx context.Context, role domain.ReleaseOperatorRole, limit, offset int) ([]domain.ReleaseOperator, int, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	args := []any{limit, offset}
	where := "WHERE TRUE"
	if role != "" {
		args = append(args, role)
		where += fmt.Sprintf(" AND role=$%d", len(args))
	}
	rows, err := p.pool.Query(ctx, fmt.Sprintf(`SELECT id,name,role FROM release_operators %s ORDER BY name LIMIT $1 OFFSET $2`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list release_operators: %w", err)
	}
	defer rows.Close()
	items := make([]domain.ReleaseOperator, 0)
	for rows.Next() {
		var item domain.ReleaseOperator
		if err := rows.Scan(&item.ID, &item.Name, &item.Role); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	countArgs := args[2:]
	countWhere := "WHERE TRUE"
	if role != "" {
		countWhere += " AND role=$1"
	}
	if err := p.pool.QueryRow(ctx, "SELECT count(*) FROM release_operators "+countWhere, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count release_operators: %w", err)
	}
	return items, total, nil
}

func (p *Postgres) RenameReleaseOperator(ctx context.Context, id, name string) error {
	return renameReleaseOperator(ctx, p.pool, id, name)
}

func (t *transaction) RenameReleaseOperator(ctx context.Context, id, name string) error {
	return renameReleaseOperator(ctx, t.tx, id, name)
}

func renameReleaseOperator(ctx context.Context, q sqler, id, name string) error {
	result, err := q.Exec(ctx, `UPDATE release_operators SET name=$1 WHERE id=$2`, name, id)
	if err != nil {
		return fmt.Errorf("rename release_operator: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrNotFound
	}
	return nil
}
