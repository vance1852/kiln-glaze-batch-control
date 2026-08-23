package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	auditpkg "firmware-rollout-control/internal/audit"
	"firmware-rollout-control/internal/repository"
)

type recordingAuditRepository struct {
	repository.Repository
	mu    sync.Mutex
	heads map[string]struct{}
	err   error
}

func (r *recordingAuditRepository) WriteAudit(_ context.Context, in repository.AuditInput) error {
	if r.err != nil {
		return r.err
	}
	var detail map[string]any
	if err := json.Unmarshal(in.Detail, &detail); err != nil {
		return err
	}
	head, _ := detail["chain_head"].(string)
	r.mu.Lock()
	r.heads[head] = struct{}{}
	r.mu.Unlock()
	return nil
}

func TestConcurrentAuditEventsGetUniqueChainHeads(t *testing.T) {
	repo := &recordingAuditRepository{heads: make(map[string]struct{})}
	svc := New(repo)
	const writers = 64
	var done sync.WaitGroup
	done.Add(writers)
	for i := range writers {
		go func(i int) {
			defer done.Done()
			event := auditpkg.Event{RequestID: fmt.Sprintf("req-%d", i), ObjectType: "managed_device_task", ObjectID: "kiln-1", Action: "firing", Outcome: "success"}
			if err := svc.WriteAuditEvent(t.Context(), event); err != nil {
				t.Errorf("write %d: %v", i, err)
			}
		}(i)
	}
	done.Wait()
	if len(repo.heads) != writers {
		t.Fatalf("unique chain heads=%d want %d", len(repo.heads), writers)
	}
}

func TestAuditChainHeadRollsBackWhenWriteFails(t *testing.T) {
	repo := &recordingAuditRepository{heads: make(map[string]struct{}), err: fmt.Errorf("persist denied")}
	svc := New(repo)
	err := svc.WriteAuditEvent(t.Context(), auditpkg.Event{RequestID: "req-1", ObjectType: "managed_device_task", ObjectID: "kiln-1", Action: "firing", Outcome: "success"})
	if err == nil {
		t.Fatal("write succeeded despite persist failure")
	}
	if svc.auditChainHead() != "" {
		t.Fatalf("chain head advanced after failed write: %s", svc.auditChainHead())
	}
	if len(repo.heads) != 0 {
		t.Fatalf("recorded heads=%d want 0", len(repo.heads))
	}
}
