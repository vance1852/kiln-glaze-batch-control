package service

import (
  "context"
  "testing"
  "time"
  "firmware-rollout-control/internal/repository"
)

type privateActivationRepo0039 struct { repository.Repository; records []repository.ActivationRecord }
func (r *privateActivationRepo0039) ListActivation(context.Context,string)([]repository.ActivationRecord,error){return r.records,nil}
func TestVerifyActivationRejectsInvalidSequence(t *testing.T){ repo:=&privateActivationRepo0039{records:[]repository.ActivationRecord{{To:"accepted",Location:"kiln",RecordedAt:time.Unix(1700000000,0).UTC()},{From:nil,To:"verified",Location:"kiln",RecordedAt:time.Unix(1700000100,0).UTC()}}}; got,err:=New(repo).VerifyActivation(t.Context(),"job-39"); if err==nil{t.Fatal("invalid activation sequence accepted")}; if got!=nil{t.Fatalf("records returned on validation error: %+v",got)}}
