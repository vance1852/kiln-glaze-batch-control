package service

import (
  "context"
  "errors"
  "testing"
  "time"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateExpiringRepo0050 struct { repository.Repository; err error }
func (r *privateExpiringRepo0050) ExpiringDeploymentJobs(context.Context,time.Time,int)([]domain.DeploymentJob,error){return nil,r.err}
func TestExpiringDeploymentJobsPropagatesQueryError(t *testing.T){want:=errors.New("expiry query down"); got,err:=New(&privateExpiringRepo0050{err:want}).ExpiringDeploymentJobs(t.Context(),1700000000,10); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if got!=nil{t.Fatalf("items returned on error: %+v",got)}}
