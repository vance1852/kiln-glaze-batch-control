package service

import (
  "context"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateAssignRepo0033 struct { repository.Repository; created bool }
func (r *privateAssignRepo0033) InTx(ctx context.Context, fn func(repository.Repository) error) error{return fn(r)}
func (r *privateAssignRepo0033) CreateAssignment(context.Context,domain.Assignment) error{r.created=true;return nil}
func (r *privateAssignRepo0033) ValidateRolloutCampaignManagedDevice(context.Context,string,string) error{return nil}
func (r *privateAssignRepo0033) WriteAudit(context.Context,repository.AuditInput) error{return nil}
func TestAssignManagedDeviceRejectsUnauthorizedRole(t *testing.T){repo:=&privateAssignRepo0033{}; a:=domain.Assignment{ID:"assign-33",RolloutCampaignID:"camp-33",ManagedDeviceID:"dev-33",ReleaseOperatorID:"op-33"}; op:=domain.ReleaseOperator{ID:"op-33",Name:"Reviewer",Role:domain.RoleQualityReviewer}; err:=New(repo).AssignManagedDevice(t.Context(),RequestMeta{RequestID:"req-33"},a,op); if err==nil{t.Fatal("unauthorized assignment succeeded")}; if repo.created{t.Fatal("unauthorized assignment persisted")}}
