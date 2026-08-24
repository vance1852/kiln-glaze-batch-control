package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateAuditRepo0044 struct { repository.Repository; err error }
func (r *privateAuditRepo0044) AuditHistory(context.Context,string,string,int)([]domain.AuditSummary,error){return nil,r.err}
func TestAuditHistoryPropagatesQueryError0044(t *testing.T){want:=errors.New("audit store unavailable"); got,err:=New(&privateAuditRepo0044{err:want}).AuditHistory(t.Context(),"deployment_job","job-42",20); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if got!=nil{t.Fatalf("records returned on error: %+v",got)}}
