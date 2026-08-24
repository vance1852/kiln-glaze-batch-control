package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/repository"
)

type privateRenameRepo0053 struct { repository.Repository; err error; audited bool }
func (r *privateRenameRepo0053) InTx(ctx context.Context, fn func(repository.Repository) error) error{return fn(r)}
func (r *privateRenameRepo0053) RenameReleaseOperator(context.Context,string,string) error{return r.err}
func (r *privateRenameRepo0053) WriteAudit(context.Context,repository.AuditInput) error{r.audited=true;return nil}
func TestRenameReleaseOperatorPropagatesWriteError(t *testing.T){want:=errors.New("rename conflict"); repo:=&privateRenameRepo0053{err:want}; err:=New(repo).RenameReleaseOperator(t.Context(),RequestMeta{RequestID:"req-53"},"op-53","Supervisor 53"); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if repo.audited{t.Fatal("audit written after failed rename")}}
