package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateOperatorListRepo0052 struct { repository.Repository; err error }
func (r *privateOperatorListRepo0052) ListReleaseOperators(context.Context,domain.ReleaseOperatorRole,int,int)([]domain.ReleaseOperator,int,error){return nil,0,r.err}
func TestListReleaseOperatorsPropagatesQueryError(t *testing.T){want:=errors.New("operator list unavailable"); got,total,err:=New(&privateOperatorListRepo0052{err:want}).ListReleaseOperators(t.Context(),domain.RoleSafetySupervisor,20,0); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if got!=nil||total!=0{t.Fatalf("operators returned on error: %v %d",got,total)}}
