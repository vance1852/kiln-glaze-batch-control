package integration

import (
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
	"github.com/google/uuid"
)

func TestAssignmentAuditFailureDoesNotLeaveAssignment(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	campaign := uuid.NewString()
	device := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO rollout_campaigns(id,code,name,status,timezone,starts_at,ends_at,created_by) VALUES($1,$2,$3,'active','UTC',$4,$5,$6)`, campaign, "KILN-ROLLBACK", "Kiln rollback", time.Now().UTC(), time.Now().UTC().Add(time.Hour), operator); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO managed_devices(id,rollout_campaign_id,code,rollout_lane,required_successes) VALUES($1,$2,$3,'north',1)`, device, campaign, "K-01"); err != nil {
		t.Fatal(err)
	}
	assignment := domain.Assignment{ID: uuid.NewString(), RolloutCampaignID: campaign, ManagedDeviceID: device, ReleaseOperatorID: operator, StartsAt: time.Now().UTC(), EndsAt: time.Now().UTC().Add(time.Hour), Status: "queued"}
	missingActor := uuid.NewString()
	err := service.New(repository.NewPostgres(pool)).AssignManagedDevice(ctx, service.RequestMeta{RequestID: "audit-failure", ReleaseOperatorID: &missingActor}, assignment, domain.ReleaseOperator{ID: operator, Name: "operator", Role: domain.RoleManagedDeviceOperator})
	if err == nil {
		t.Fatal("assignment succeeded despite rejected audit event")
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM assignments WHERE id=$1`, assignment.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("assignment survived audit rollback: %d", count)
	}
}
