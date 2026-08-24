package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateOperatorRepo0051 struct { repository.Repository; err error }
func (r *privateOperatorRepo0051) GetReleaseOperator(context.Context,string)(domain.ReleaseOperator,error){return domain.ReleaseOperator{},r.err}
func TestLoadReleaseOperatorPropagatesNotFound(t *testing.T){got,err:=New(&privateOperatorRepo0051{err:domain.ErrNotFound}).LoadReleaseOperator(t.Context(),"op-51"); if !errors.Is(err,domain.ErrNotFound){t.Fatalf("err=%v",err)}; if got.ID!=""{t.Fatalf("operator returned on error: %+v",got)}}
