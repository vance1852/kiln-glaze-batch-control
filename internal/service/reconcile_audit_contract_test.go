package service

import (
	"context"
	"testing"
	"time"

	"firmware-rollout-control/internal/repository"
)

type reconcileAuditRepository struct {
	repository.Repository
	audited   bool
	unaudited bool
}

func (r *reconcileAuditRepository) MarkExpiredDeploymentJobs(context.Context, time.Time, int) (repository.ReconcileResult, error) {
	r.audited = true
	return repository.ReconcileResult{Scanned: 1, Marked: 1}, nil
}

func (r *reconcileAuditRepository) MarkExpiredDeploymentJobsWithoutAudit(context.Context, time.Time, int) (repository.ReconcileResult, error) {
	r.unaudited = true
	return repository.ReconcileResult{Scanned: 1, Marked: 1}, nil
}

func TestExpiredDeploymentJobReconciliationUsesAuditedWrite(t *testing.T) {
	repo := &reconcileAuditRepository{}
	result, err := New(repo).MarkExpiredDeploymentJobs(t.Context(), time.Now().UTC(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.Marked != 1 || !repo.audited || repo.unaudited {
		t.Fatalf("result=%+v audited=%v unaudited=%v", result, repo.audited, repo.unaudited)
	}
}
