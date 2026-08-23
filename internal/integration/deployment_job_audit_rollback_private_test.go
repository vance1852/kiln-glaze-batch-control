package integration

import (
	"testing"
	"time"

	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
	"github.com/google/uuid"
)

func TestDeploymentJobAuditFailureRollsBackCreation(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operatorID := insertReleaseOperator(t, ctx, pool, "safety_supervisor")
	campaignID := uuid.NewString()
	deviceID := uuid.NewString()
	now := time.Now().UTC()
	if _, err := pool.Exec(ctx, `INSERT INTO rollout_campaigns(id,code,name,status,timezone,starts_at,ends_at,created_by) VALUES($1,$2,$3,'active','UTC',$4,$5,$6)`, campaignID, "KILN-JOB-ROLLBACK", "Audit rollback", now, now.Add(2*time.Hour), operatorID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO managed_devices(id,rollout_campaign_id,code,rollout_lane,required_successes) VALUES($1,$2,$3,'firing',1)`, deviceID, campaignID, "KILN-03"); err != nil {
		t.Fatal(err)
	}
	missingActor := uuid.NewString()
	_, err := service.New(repository.NewPostgres(pool)).CreateDeploymentJob(ctx, service.RequestMeta{RequestID: "job-audit-rollback", ReleaseOperatorID: &missingActor}, repository.DeploymentJobInput{RolloutCampaignID: campaignID, ManagedDeviceID: deviceID, TaskCode: "FIRE-03", ExpiresAt: now.Add(time.Hour)})
	if err == nil {
		t.Fatal("job creation succeeded despite rejected audit event")
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM deployment_jobs WHERE task_code='FIRE-03'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("failed job creation left %d persistent row", count)
	}
	created, err := service.New(repository.NewPostgres(pool)).CreateDeploymentJob(ctx, service.RequestMeta{RequestID: "job-audit-valid", ReleaseOperatorID: &operatorID}, repository.DeploymentJobInput{RolloutCampaignID: campaignID, ManagedDeviceID: deviceID, TaskCode: "FIRE-03-RETRY", ExpiresAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatalf("valid retry failed: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_id=$1 AND action='create'`, created.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("valid retry audit count=%d", count)
	}
}
