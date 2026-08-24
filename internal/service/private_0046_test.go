package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateDeviceListRepo0046 struct { repository.Repository; err error }
func (r *privateDeviceListRepo0046) ListRolloutCampaignManagedDevices(context.Context,string)([]domain.ManagedDevice,error){return nil,r.err}
func TestListCampaignManagedDevicesPropagatesQueryError(t *testing.T){want:=errors.New("device query down"); got,err:=New(&privateDeviceListRepo0046{err:want}).ListRolloutCampaignManagedDevices(t.Context(),"camp-46"); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if got!=nil{t.Fatalf("devices returned on error: %+v",got)}}
