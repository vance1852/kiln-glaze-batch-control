package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/repository"
)

type privateAlertRepo0035 struct { repository.Repository; err error; audited bool }
func (r *privateAlertRepo0035) InTx(ctx context.Context, fn func(repository.Repository) error) error{return fn(r)}
func (r *privateAlertRepo0035) CloseHealthAlert(context.Context,string) error{return r.err}
func (r *privateAlertRepo0035) WriteAudit(context.Context,repository.AuditInput) error{r.audited=true;return nil}
func TestCloseHealthAlertPropagatesRepositoryError(t *testing.T){want:=errors.New("lock conflict"); repo:=&privateAlertRepo0035{err:want}; err:=New(repo).CloseHealthAlert(t.Context(),RequestMeta{RequestID:"req-34"},"alert-34"); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if repo.audited{t.Fatal("success audit written after failed transition")}}
