package repository

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

func (p *Postgres) CreateDeploymentJobsBulk(ctx context.Context, inputs []DeploymentJobInput) ([]domain.DeploymentJob, error) {
	var deployment_jobs []domain.DeploymentJob
	err := p.InTx(ctx, func(tx Repository) error {
		for _, input := range inputs {
			task, err := tx.CreateDeploymentJob(ctx, input)
			if err != nil {
				return fmt.Errorf("create task %s: %w", input.TaskCode, err)
			}
			deployment_jobs = append(deployment_jobs, task)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return append([]domain.DeploymentJob(nil), deployment_jobs...), nil
}

func ValidateBulkCapacity(inputs []DeploymentJobInput, capacity int) error {
	if capacity < 1 || len(inputs) > capacity {
		return domain.ErrCapacityExceeded
	}
	return nil
}
