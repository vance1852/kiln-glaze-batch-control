package integration

import (
	"testing"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
	"github.com/google/uuid"
)

func TestReleaseOperatorCreationRollsBackWhenAuditWriteFails(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	svc := service.New(repository.NewPostgres(pool))
	missingReleaseOperator := uuid.NewString()
	_, err := svc.RegisterReleaseOperator(ctx, service.RequestMeta{RequestID: "rollback", ReleaseOperatorID: &missingReleaseOperator}, "Rollback ReleaseOperator", domain.RoleSafetySupervisor)
	if err == nil {
		t.Fatal("release_operator creation succeeded despite rejected audit event")
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM release_operators WHERE name='Rollback ReleaseOperator'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("release_operator survived audit rollback: %d", count)
	}
}

func TestRolloutWaveTransitionRollsBackWhenAuditWriteFails(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	repo := repository.NewPostgres(pool)
	rollout_waveID, err := repo.CreateRolloutWave(ctx, repository.RolloutWaveInput{Code: "ROLLBACK-ROUND", Method: "daily-managed_device", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	missingReleaseOperator := uuid.NewString()
	err = service.New(repo).StartRolloutWave(ctx, service.RequestMeta{RequestID: "rollback", ReleaseOperatorID: &missingReleaseOperator}, rollout_waveID, 1)
	if err == nil {
		t.Fatal("rollout_wave transition succeeded despite rejected audit event")
	}
	var status string
	var version int64
	if err := pool.QueryRow(ctx, `SELECT status,version FROM rollout_waves WHERE id=$1`, rollout_waveID).Scan(&status, &version); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.RolloutWaveQueued) || version != 1 {
		t.Fatalf("rollout_wave survived audit rollback: status=%s version=%d", status, version)
	}
}
