package integration

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"firmware-rollout-control/internal/db"
	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
	"github.com/google/uuid"
)

func openDatabase(t *testing.T) (*db.Pool, context.Context) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx := context.Background()
	pool, err := db.Open(ctx, url, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx, pool); err != nil {
		pool.Close()
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE TABLE audit_events, health_alerts, installation_reports, rollout_wave_items, rollout_waves, activation_events, deployment_jobs, managed_devices, rollout_campaigns, release_operators, idempotency_keys CASCADE`); err != nil {
		pool.Close()
		t.Fatal(err)
	}
	return pool, ctx
}

func insertReleaseOperator(t *testing.T, ctx context.Context, pool *db.Pool, role string) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO release_operators(id,name,role) VALUES ($1,$2,$3)`, id, "Test ReleaseOperator", role); err != nil {
		t.Fatal(err)
	}
	return id
}

func createWorkflow(t *testing.T, ctx context.Context, svc *service.Service, release_operator string) (service.CreateRolloutCampaignResponse, domain.DeploymentJob) {
	t.Helper()
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	svc.WithClock(func() time.Time { return now })
	rollout_campaign, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: "req-create"}, service.CreateRolloutCampaignRequest{
		Code: "PLAN-001", Name: "North river survey", Timezone: "Asia/Shanghai", StartsAt: now, EndsAt: now.Add(24 * time.Hour), CreatedBy: release_operator,
		ManagedDevices: []repository.ManagedDeviceInput{{Code: "N-01", RolloutLane: "A-101", RequiredSuccesses: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ScheduleRolloutCampaign(ctx, service.RequestMeta{RequestID: "req-schedule"}, rollout_campaign.RolloutCampaign.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.ActivateRolloutCampaign(ctx, service.RequestMeta{RequestID: "req-collect"}, rollout_campaign.RolloutCampaign.ID, 2); err != nil {
		t.Fatal(err)
	}
	task, err := svc.CreateDeploymentJob(ctx, service.RequestMeta{RequestID: "req-task"}, repository.DeploymentJobInput{RolloutCampaignID: rollout_campaign.RolloutCampaign.ID, ManagedDeviceID: rollout_campaign.ManagedDeviceIDs[0], TaskCode: "S-001", ExpiresAt: now.Add(12 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteDeploymentJob(ctx, service.RequestMeta{RequestID: "req-collect-task"}, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.ActivationDeploymentJob(ctx, service.RequestMeta{RequestID: "req-activation"}, repository.ActivationInput{DeploymentJobID: task.ID, To: release_operator, Location: "A-101", RecordedAt: now}, 2); err != nil {
		t.Fatal(err)
	}
	if err := svc.AcceptDeploymentJob(ctx, service.RequestMeta{RequestID: "req-receive"}, repository.ActivationInput{DeploymentJobID: task.ID, To: release_operator, Location: "ManagedDevice bay", RecordedAt: now}, 3); err != nil {
		t.Fatal(err)
	}
	page, err := svc.ListDeploymentJobs(ctx, 0, 10, rollout_campaign.RolloutCampaign.ID, domain.DeploymentJobAccepted)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("task listing = %+v, %v", page, err)
	}
	return rollout_campaign, page.Items[0]
}

func TestDeploymentJobWorkflowPersistsAcrossOperations(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	rollout_campaign, task := createWorkflow(t, ctx, svc, release_operator)
	if rollout_campaign.RolloutCampaign.Status != domain.RolloutCampaignDraft {
		t.Fatalf("created rollout_campaign status = %s", rollout_campaign.RolloutCampaign.Status)
	}
	if task.Status != domain.DeploymentJobAccepted {
		t.Fatalf("task status = %s", task.Status)
	}
	if task.Version != 4 {
		t.Fatalf("task version = %d", task.Version)
	}
	var auditCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_id=$1`, task.ID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 4 {
		t.Fatalf("audit count = %d, want 4", auditCount)
	}
}

func TestRolloutCampaignCreationRollsBackWhenASecondManagedDeviceViolatesConstraint(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "safety_supervisor")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	_, err := svc.CreateRolloutCampaign(ctx, service.RequestMeta{RequestID: "req-rollback"}, service.CreateRolloutCampaignRequest{
		Code: "PLAN-ROLLBACK", Name: "Rollback test", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: release_operator,
		ManagedDevices: []repository.ManagedDeviceInput{{Code: "GOOD", RolloutLane: "A", RequiredSuccesses: 1}, {Code: "BAD", RolloutLane: "B", RequiredSuccesses: 0}},
	})
	if err == nil {
		t.Fatal("invalid second managed_device was accepted")
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM rollout_campaigns WHERE code='PLAN-ROLLBACK'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollout_campaign survived rollback: %d", count)
	}
}

func TestConcurrentDeploymentJobTransitionAllowsOnlyOneVersionWinner(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	repo := repository.NewPostgres(pool)
	rollout_campaign, task := createWorkflow(t, ctx, svc, release_operator)
	_ = rollout_campaign
	// The workflow leaves the task at received/version 4. Two workers race to start testing.
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- repo.MoveDeploymentJob(ctx, task.ID, domain.DeploymentJobInProgress, 4, time.Now().UTC())
		}()
	}
	wg.Wait()
	close(results)
	var success, conflict int
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, domain.ErrConflict) {
			conflict++
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}

func TestMigrationAndStateSurviveDatabaseReopen(t *testing.T) {
	pool, ctx := openDatabase(t)
	release_operator := insertReleaseOperator(t, ctx, pool, "managed_device_operator")
	svc := service.New(repository.NewPostgres(pool))
	_, task := createWorkflow(t, ctx, svc, release_operator)
	pool.Close()
	url := os.Getenv("DATABASE_URL")
	reopened, err := db.Open(ctx, url, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := repository.NewPostgres(reopened).GetDeploymentJob(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TaskCode != "S-001" || loaded.Status != domain.DeploymentJobAccepted {
		t.Fatalf("reopened task = %+v", loaded)
	}
}
