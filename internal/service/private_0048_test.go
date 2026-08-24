package service

import (
  "context"
  "errors"
  "testing"
  "time"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateComplianceRepo0048 struct { repository.Repository; err error }
func (r *privateComplianceRepo0048) GetRolloutCampaign(context.Context,string)(domain.RolloutCampaign,error){return domain.RolloutCampaign{ID:"camp-47"},nil}
func (r *privateComplianceRepo0048) ComplianceReport(context.Context,string,time.Time)(domain.ComplianceReport,error){return domain.ComplianceReport{},r.err}
func TestComplianceReportPropagatesQueryError0048(t *testing.T){want:=errors.New("aggregate unavailable"); got,err:=New(&privateComplianceRepo0048{err:want}).ComplianceReport(t.Context(),"camp-47"); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if got.RolloutCampaignID != "" || got.OpenHealthAlerts != 0 {t.Fatalf("empty report returned as success: %+v",got)}}
