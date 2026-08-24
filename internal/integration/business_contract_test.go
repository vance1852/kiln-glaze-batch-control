package integration

import (
	"errors"
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
)

func TestRolloutCampaignProgressCountsEachManagedDeviceRequirementOnce(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	rollout_campaign, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: "progress"}, service.CreateRolloutCampaignRequest{
		Code: "PROGRESS", Name: "Progress", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: release_operator,
		ManagedDevices: []repository.ManagedDeviceInput{{Code: "P-1", RolloutLane: "A", RequiredSuccesses: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewPostgres(pool)
	for _, code := range []string{"P-S1", "P-S2"} {
		if _, err := repo.CreateDeploymentJob(ctx, repository.DeploymentJobInput{RolloutCampaignID: rollout_campaign.RolloutCampaign.ID, ManagedDeviceID: rollout_campaign.ManagedDeviceIDs[0], TaskCode: code, ExpiresAt: now.Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	progress, err := repo.RolloutCampaignProgress(ctx, rollout_campaign.RolloutCampaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.ManagedDevices != 1 || progress.Required != 2 {
		t.Fatalf("progress=%+v", progress)
	}
}

func TestComplianceReportContainsOnlyItsRolloutCampaignDeploymentJobs(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	create := func(code string) service.CreateRolloutCampaignResponse {
		rollout_campaign, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: code}, service.CreateRolloutCampaignRequest{
			Code: code, Name: code, Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: release_operator,
			ManagedDevices: []repository.ManagedDeviceInput{{Code: code + "-MANAGED_DEVICE", RolloutLane: "A", RequiredSuccesses: 1}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return rollout_campaign
	}
	first, second := create("REPORT-A"), create("REPORT-B")
	repo := repository.NewPostgres(pool)
	for _, item := range []struct{ rollout_campaign, managed_device, code string }{{first.RolloutCampaign.ID, first.ManagedDeviceIDs[0], "REPORT-S1"}, {second.RolloutCampaign.ID, second.ManagedDeviceIDs[0], "REPORT-S2"}} {
		if _, err := repo.CreateDeploymentJob(ctx, repository.DeploymentJobInput{RolloutCampaignID: item.rollout_campaign, ManagedDeviceID: item.managed_device, TaskCode: item.code, ExpiresAt: now.Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	report, err := repo.ComplianceReport(ctx, first.RolloutCampaign.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Expiring) != 1 || report.Expiring[0].RolloutCampaignID != first.RolloutCampaign.ID {
		t.Fatalf("expiring=%+v", report.Expiring)
	}
}

func TestRolloutWaveRejectsDeploymentJobsThatAreNotReadyForInProgress(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	rollout_campaign, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: "rollout_wave"}, service.CreateRolloutCampaignRequest{
		Code: "ROUND-READY", Name: "RolloutWave", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: release_operator,
		ManagedDevices: []repository.ManagedDeviceInput{{Code: "B-1", RolloutLane: "A", RequiredSuccesses: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.CreateDeploymentJob(ctx, service.RequestMeta{RequestID: "task"}, repository.DeploymentJobInput{RolloutCampaignID: rollout_campaign.RolloutCampaign.ID, ManagedDeviceID: rollout_campaign.ManagedDeviceIDs[0], TaskCode: "B-S1", ExpiresAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateRolloutWave(ctx, service.RequestMeta{RequestID: "rollout_wave-create"}, repository.RolloutWaveInput{Code: "B-1", Method: "daily-managed_device", Capacity: 1}, []string{task.ID})
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("err=%v", err)
	}
	var rollout_waves int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM rollout_waves WHERE code='B-1'`).Scan(&rollout_waves); err != nil {
		t.Fatal(err)
	}
	if rollout_waves != 0 {
		t.Fatalf("partial rollout_wave count=%d", rollout_waves)
	}
}

func TestRolloutWaveAndAssignmentRejectSkippedStates(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	repo := repository.NewPostgres(pool)
	rollout_waveID, err := repo.CreateRolloutWave(ctx, repository.RolloutWaveInput{Code: "STATE-B", Method: "daily-managed_device", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CompleteRolloutWave(ctx, rollout_waveID, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("rollout_wave err=%v", err)
	}
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repo)
	now := time.Now().UTC()
	rollout_campaign, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: "assignment"}, service.CreateRolloutCampaignRequest{
		Code: "ASSIGN-STATE", Name: "Assignment", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: release_operator,
		ManagedDevices: []repository.ManagedDeviceInput{{Code: "A-1", RolloutLane: "A", RequiredSuccesses: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assignment := repository.NewAssignment(rollout_campaign.RolloutCampaign.ID, rollout_campaign.ManagedDeviceIDs[0], release_operator, now, now.Add(time.Hour))
	if err := repo.CreateAssignment(ctx, assignment); err != nil {
		t.Fatal(err)
	}
	if err := repo.AdvanceAssignment(ctx, assignment.ID, "completed", 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("assignment err=%v", err)
	}
}

func TestDeploymentJobRejectsManagedDeviceFromAnotherRolloutCampaign(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	create := func(code string) service.CreateRolloutCampaignResponse {
		rollout_campaign, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: code}, service.CreateRolloutCampaignRequest{
			Code: code, Name: code, Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: release_operator,
			ManagedDevices: []repository.ManagedDeviceInput{{Code: code + "-MANAGED_DEVICE", RolloutLane: "A", RequiredSuccesses: 1}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return rollout_campaign
	}
	first := create("PLAN-A")
	second := create("PLAN-B")
	_, err := svc.CreateDeploymentJob(ctx, service.RequestMeta{RequestID: "mismatch"}, repository.DeploymentJobInput{
		RolloutCampaignID: first.RolloutCampaign.ID, ManagedDeviceID: second.ManagedDeviceIDs[0], TaskCode: "MISMATCH-01", ExpiresAt: now.Add(time.Hour),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM deployment_jobs WHERE task_code='MISMATCH-01'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("mismatched task persisted: %d", count)
	}
}

func TestExpiredDeploymentJobReconciliationWritesAudit(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	rollout_campaign, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: "expiry-rollout_campaign"}, service.CreateRolloutCampaignRequest{
		Code: "EXPIRY-PLAN", Name: "Expiry", Timezone: "UTC", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour), CreatedBy: release_operator,
		ManagedDevices: []repository.ManagedDeviceInput{{Code: "EXPIRY-MANAGED_DEVICE", RolloutLane: "A", RequiredSuccesses: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var taskID string
	if err := pool.QueryRow(ctx, `INSERT INTO deployment_jobs(id,rollout_campaign_id,managed_device_id,task_code,status,expires_at,version) VALUES (gen_random_uuid(),$1,$2,'EXPIRED-01','queued',$3,1) RETURNING id`, rollout_campaign.RolloutCampaign.ID, rollout_campaign.ManagedDeviceIDs[0], now.Add(-time.Minute)).Scan(&taskID); err != nil {
		t.Fatal(err)
	}
	result, err := svc.MarkExpiredDeploymentJobs(ctx, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.Marked != 1 || result.Failed != 0 {
		t.Fatalf("result=%+v", result)
	}
	var status string
	var audits int
	if err := pool.QueryRow(ctx, `SELECT status FROM deployment_jobs WHERE id=$1`, taskID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_id=$1 AND action='expire'`, taskID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.DeploymentJobRejected) || audits != 1 {
		t.Fatalf("status=%s audits=%d", status, audits)
	}
}
