package service

import (
  "context"
  "errors"
  "testing"
  "time"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateReviewRepo0056 struct { repository.Repository; reviewed, moved, audited bool }
func (r *privateReviewRepo0056) InTx(ctx context.Context, fn func(repository.Repository) error) error { return fn(r) }
func (r *privateReviewRepo0056) InstallationReportTaskID(context.Context, string) (string, error) { return "task-owned-by-report", nil }
func (r *privateReviewRepo0056) ReviewInstallationReportRecord(context.Context, string, bool, int64, time.Time) error { r.reviewed = true; return nil }
func (r *privateReviewRepo0056) MoveDeploymentJob(context.Context, string, domain.DeploymentJobStatus, int64, time.Time) error { r.moved = true; return nil }
func (r *privateReviewRepo0056) CreateHealthAlert(context.Context, repository.HealthAlertInput) (string, error) { return "alert-56", nil }
func (r *privateReviewRepo0056) WriteAudit(context.Context, repository.AuditInput) error { r.audited = true; return nil }
func TestReviewInstallationReportRejectsCrossTaskIdentity(t *testing.T) {
  repo := &privateReviewRepo0056{}
  err := New(repo).ReviewInstallationReport(context.Background(), RequestMeta{RequestID: "review-56"}, "report-56", "different-task", true, 1, 1)
  if err == nil || !errors.Is(err, domain.ErrConflict) { t.Fatalf("err=%v", err) }
  if repo.reviewed || repo.moved || repo.audited { t.Fatal("cross-task review mutated state") }
}
