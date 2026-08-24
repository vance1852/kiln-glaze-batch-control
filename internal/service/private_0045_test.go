package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateCampaignListRepo0045 struct { repository.Repository; err error }
func (r *privateCampaignListRepo0045) ListRolloutCampaigns(context.Context,repository.RolloutCampaignFilter)([]domain.RolloutCampaign,int,error){return nil,0,r.err}
func TestListRolloutCampaignsPropagatesQueryError(t *testing.T){want:=errors.New("campaign query down"); got,total,err:=New(&privateCampaignListRepo0045{err:want}).ListRolloutCampaigns(t.Context(),repository.RolloutCampaignFilter{Search:"ancient kiln"}); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if got!=nil||total!=0{t.Fatalf("items=%v total=%d",got,total)}}
