package service

import (
  "context"
  "testing"
  "time"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateReportRepo0025 struct { repository.Repository; created bool; nextRole domain.ReleaseOperatorRole }
func (r *privateReportRepo0025) InTx(ctx context.Context, fn func(repository.Repository) error) error { return fn(r) }
func (r *privateReportRepo0025) ValidateInstallationReportTarget(context.Context,string,string) error { return nil }
func (r *privateReportRepo0025) GetReleaseOperator(context.Context,string) (domain.ReleaseOperator,error) { return domain.ReleaseOperator{ID:"reviewer-1",Name:"Reviewer",Role:r.nextRole},nil }
func (r *privateReportRepo0025) CreateInstallationReport(context.Context,repository.InstallationReportInput)(string,error) { r.created=true; return "report-1",nil }
func (r *privateReportRepo0025) WriteAudit(context.Context,repository.AuditInput) error { return nil }
func reportInput0025() repository.InstallationReportInput { return repository.InstallationReportInput{DeploymentJobID:"job-1",RolloutWaveID:"wave-1",RecorderID:"reviewer-1",ObservedAt:time.Unix(1700000000,0).UTC(),RiskScore:0.2,AlertThreshold:0.8,Scale:"0-1"} }
func TestSubmitInstallationReportRejectsQualityReviewer(t *testing.T) {
  repo:=&privateReportRepo0025{nextRole:domain.RoleQualityReviewer}; _,err:=New(repo).SubmitInstallationReport(t.Context(),RequestMeta{RequestID:"req-25"},reportInput0025()); if err==nil { t.Fatal("quality reviewer was allowed to submit installation report") }; if repo.created { t.Fatal("unauthorized report was persisted") }
  repo=&privateReportRepo0025{nextRole:domain.RoleInstallationOperator}; id,err:=New(repo).SubmitInstallationReport(t.Context(),RequestMeta{RequestID:"req-25b"},reportInput0025()); if err!=nil || id!="report-1" { t.Fatalf("authorized submission id=%q err=%v",id,err) }
}
