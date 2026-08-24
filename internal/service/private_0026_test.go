package service

import (
  "context"
  "testing"
  "time"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateReviewRepo0026 struct { repository.Repository; alerts int; moves int }
func (r *privateReviewRepo0026) InTx(ctx context.Context, fn func(repository.Repository) error) error { return fn(r) }
func (r *privateReviewRepo0026) InstallationReportTaskID(context.Context,string)(string,error){ return "job-26",nil }
func (r *privateReviewRepo0026) ReviewInstallationReportRecord(context.Context,string,bool,int64,time.Time) error { return nil }
func (r *privateReviewRepo0026) MoveDeploymentJob(context.Context,string,domain.DeploymentJobStatus,int64,time.Time) error { r.moves++; return nil }
func (r *privateReviewRepo0026) CreateHealthAlert(context.Context,repository.HealthAlertInput)(string,error){ r.alerts++; return "alert-26",nil }
func (r *privateReviewRepo0026) WriteAudit(context.Context,repository.AuditInput) error { return nil }
func TestAcceptedInstallationReportDoesNotOpenSafetyAlert(t *testing.T){ repo:=&privateReviewRepo0026{}; err:=New(repo).WithClock(func()time.Time{return time.Unix(1700000000,0).UTC()}).ReviewInstallationReport(t.Context(),RequestMeta{RequestID:"req-26"},"report-26","job-26",true,1,1); if err!=nil { t.Fatal(err) }; if repo.alerts!=0 { t.Fatalf("accepted report opened %d alerts",repo.alerts) }; if repo.moves!=1 { t.Fatalf("task transition count=%d",repo.moves) } }
