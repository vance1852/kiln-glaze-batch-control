package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/repository"
)

type privateAdvanceRepo0037 struct { repository.Repository; err error; audited bool }
func (r *privateAdvanceRepo0037) InTx(ctx context.Context, fn func(repository.Repository) error) error{return fn(r)}
func (r *privateAdvanceRepo0037) AdvanceAssignment(context.Context,string,string,int64) error{return r.err}
func (r *privateAdvanceRepo0037) WriteAudit(context.Context,repository.AuditInput) error{r.audited=true;return nil}
func TestAdvanceAssignmentPropagatesRepositoryError0037(t *testing.T){want:=errors.New("version conflict"); repo:=&privateAdvanceRepo0037{err:want}; err:=New(repo).AdvanceAssignment(t.Context(),RequestMeta{RequestID:"req-36"},"assign-36","active",1); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if repo.audited{t.Fatal("success audit written after failed advance")}}
