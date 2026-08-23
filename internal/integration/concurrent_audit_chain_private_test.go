package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	auditpkg "firmware-rollout-control/internal/audit"
	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
	"github.com/google/uuid"
)

func TestConcurrentAuditEventsKeepUniqueChainHeads(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	svc := service.New(repository.NewPostgres(pool))

	const events = 24
	objectIDs := make([]string, events)
	for index := range objectIDs {
		objectIDs[index] = uuid.NewString()
	}
	start := make(chan struct{})
	errors := make(chan error, events)
	var writers sync.WaitGroup
	for index := 0; index < events; index++ {
		index := index
		writers.Add(1)
		go func() {
			defer writers.Done()
			<-start
			err := svc.WriteAuditEvent(ctx, auditpkg.Event{
				RequestID:  fmt.Sprintf("kiln-chain-%02d", index),
				ObjectType: "kiln_firing_audit_0007",
				ObjectID:   objectIDs[index],
				Action:     "record_temperature_curve",
				Outcome:    "success",
				CreatedAt:  time.Unix(int64(1000+index), 0).UTC(),
			})
			if err != nil {
				errors <- err
			}
		}()
	}
	close(start)
	writers.Wait()
	close(errors)
	for err := range errors {
		t.Fatalf("concurrent audit write failed: %v", err)
	}

	var stored, uniqueHeads int
	if err := pool.QueryRow(ctx, `SELECT count(*),count(DISTINCT detail->>'chain_head') FROM audit_events WHERE object_type='kiln_firing_audit_0007'`).Scan(&stored, &uniqueHeads); err != nil {
		t.Fatal(err)
	}
	if stored != events || uniqueHeads != events {
		t.Fatalf("stored events=%d unique chain heads=%d, want %d of each", stored, uniqueHeads, events)
	}
}
