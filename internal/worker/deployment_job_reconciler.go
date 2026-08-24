package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"firmware-rollout-control/internal/domain"
)

type DeploymentJobExpiryRepository interface {
	ExpiringDeploymentJobs(context.Context, time.Time, int) ([]domain.DeploymentJob, error)
}

type DeploymentJobExpiryReconciler struct {
	repo    DeploymentJobExpiryRepository
	log     *slog.Logger
	metrics *Metrics
}

func NewDeploymentJobExpiryReconciler(repo DeploymentJobExpiryRepository, logger *slog.Logger, metrics *Metrics) *DeploymentJobExpiryReconciler {
	if logger == nil {
		logger = slog.Default()
	}
	if metrics == nil {
		metrics = &Metrics{}
	}
	return &DeploymentJobExpiryReconciler{repo: repo, log: logger, metrics: metrics}
}

func (r *DeploymentJobExpiryReconciler) Reconcile(ctx context.Context, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.repo == nil {
		return fmt.Errorf("task expiry repository is nil")
	}
	r.metrics.RecordRun()
	items, err := r.repo.ExpiringDeploymentJobs(ctx, now, 100)
	if err != nil {
		r.metrics.RecordFailure()
		return err
	}
	r.metrics.RecordDue(len(items))
	for _, item := range items {
		r.log.Warn("task is near expiry", "deployment_job_id", item.ID, "task_code", item.TaskCode, "expires_at", item.ExpiresAt)
	}
	return nil
}
