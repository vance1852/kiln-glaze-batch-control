package repository

import (
	"context"
	"fmt"
)

func (p *Postgres) CheckConstraints(ctx context.Context) error {
	checks := []struct{ name, query string }{
		{"schema_migrations", `SELECT 1 FROM schema_migrations LIMIT 1`},
		{"rollout_campaigns", `SELECT 1 FROM rollout_campaigns LIMIT 1`},
		{"deployment_jobs", `SELECT 1 FROM deployment_jobs LIMIT 1`},
		{"audit_events", `SELECT 1 FROM audit_events LIMIT 1`},
	}
	for _, check := range checks {
		if _, err := p.pool.Exec(ctx, check.query); err != nil {
			return fmt.Errorf("constraint check %s: %w", check.name, err)
		}
	}
	return nil
}
