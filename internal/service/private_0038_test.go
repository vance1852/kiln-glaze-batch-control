package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/repository"
)

type privateWaveRepo0038 struct { repository.Repository; err error; audited bool }
func (r *privateWaveRepo0038) InTx(ctx context.Context, fn func(repository.Repository) error) error{return fn(r)}
func (r *privateWaveRepo0038) StartRolloutWave(context.Context,string,int64) error{return r.err}
func (r *privateWaveRepo0038) CompleteRolloutWave(context.Context,string,int64) error{return r.err}
func (r *privateWaveRepo0038) CancelRolloutWave(context.Context,string,int64) error{return r.err}
func (r *privateWaveRepo0038) WriteAudit(context.Context,repository.AuditInput) error{r.audited=true;return nil}
func TestStartRolloutWavePropagatesTransitionError(t *testing.T){want:=errors.New("wave conflict"); repo:=&privateWaveRepo0038{err:want}; err:=New(repo).StartRolloutWave(t.Context(),RequestMeta{RequestID:"req-38"},"wave-38",1); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if repo.audited{t.Fatal("audit written after failed transition")}}
