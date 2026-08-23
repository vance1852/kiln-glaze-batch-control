package integration

import (
	"errors"
	"testing"

	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
	"github.com/google/uuid"
)

func TestFailedIdempotencyClaimCanRetryAfterRestart(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	repo := repository.NewPostgres(pool)
	key := "kiln-create-" + uuid.NewString()
	body := []byte(`{"kiln":"K-08","curve":"reduction"}`)

	first := service.New(repo)
	firstStore := service.NewPersistentIdempotency(repo)
	_, _, err := first.IdempotentCreate(ctx, firstStore, key, body, func() (int, any, error) {
		return 0, nil, errors.New("firing plan storage unavailable")
	})
	if err == nil {
		t.Fatal("failed create unexpectedly succeeded")
	}

	restarted := service.New(repo)
	restartedStore := service.NewPersistentIdempotency(repo)
	createCalls := 0
	code, response, err := restarted.IdempotentCreate(ctx, restartedStore, key, body, func() (int, any, error) {
		createCalls++
		return 201, map[string]string{"job_id": "firing-08"}, nil
	})
	if err != nil {
		t.Fatalf("retry after service restart failed: %v", err)
	}
	if createCalls != 1 || code != 201 || string(response) != `{"job_id":"firing-08"}` {
		t.Fatalf("retry calls=%d code=%d response=%s", createCalls, code, response)
	}

	replayCalls := 0
	replayCode, replay, err := restarted.IdempotentCreate(ctx, restartedStore, key, body, func() (int, any, error) {
		replayCalls++
		return 500, nil, errors.New("replay must not create")
	})
	if err != nil || replayCalls != 0 || replayCode != 201 || string(replay) != string(response) {
		t.Fatalf("completed replay calls=%d code=%d response=%s err=%v", replayCalls, replayCode, replay, err)
	}
}
