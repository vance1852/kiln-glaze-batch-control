package service

import (
  "context"
  "testing"
  "time"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateCampaignRepo0029 struct { repository.Repository; campaigns int }
func (r *privateCampaignRepo0029) InTx(ctx context.Context, fn func(repository.Repository) error) error{return fn(r)}
func (r *privateCampaignRepo0029) CreateRolloutCampaign(context.Context,*domain.RolloutCampaign) error{r.campaigns++;return nil}
func (r *privateCampaignRepo0029) CreateManagedDevice(context.Context,repository.ManagedDeviceInput)(string,error){return "md-29",nil}
func (r *privateCampaignRepo0029) WriteAudit(context.Context,repository.AuditInput) error{return nil}
func TestCreateRolloutCampaignRejectsDuplicateManagedDeviceCodes(t *testing.T){repo:=&privateCampaignRepo0029{}; now:=time.Unix(1700000000,0).UTC(); in:=CreateRolloutCampaignRequest{Code:"CAMP_29",Name:"Kiln Campaign",Timezone:"UTC",StartsAt:now.Add(time.Hour),EndsAt:now.Add(2*time.Hour),CreatedBy:"operator-29",ManagedDevices:[]repository.ManagedDeviceInput{{Code:"KILN_29",RolloutLane:"NORTH",RequiredSuccesses:1},{Code:"KILN_29",RolloutLane:"SOUTH",RequiredSuccesses:1}}}; _,err:=New(repo).WithClock(func()time.Time{return now}).CreateRolloutCampaign(t.Context(),RequestMeta{RequestID:"req-29"},in); if err==nil{t.Fatal("duplicate managed device code was accepted")}; if repo.campaigns!=0{t.Fatalf("campaign writes=%d",repo.campaigns)}}
