package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateSummaryRepo0041 struct { repository.Repository; err error }
func (r *privateSummaryRepo0041) GetRolloutCampaign(context.Context,string)(domain.RolloutCampaign,error){return domain.RolloutCampaign{ID:"camp-41"},nil}
func (r *privateSummaryRepo0041) RolloutCampaignProgress(context.Context,string)(domain.RolloutCampaignProgress,error){return domain.RolloutCampaignProgress{},r.err}
func TestRolloutCampaignSummaryPropagatesProgressError(t *testing.T){want:=errors.New("progress query failed"); repo:=&privateSummaryRepo0041{err:want}; got,err:=New(repo).RolloutCampaignSummary(t.Context(),"camp-41"); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if got.RolloutCampaign.ID!=""{t.Fatalf("partial summary returned: %+v",got)}}
