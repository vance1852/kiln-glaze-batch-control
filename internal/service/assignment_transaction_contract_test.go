package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

// assignmentTransactionRepository simulates InTx with a pending buffer so that
// writes performed through the transaction are only committed on commit, while
// writes performed directly on the repository (the pre-fix bug) bypass the
// transaction and persist immediately.
type assignmentTransactionRepository struct {
	repository.Repository
	pending  map[string]domain.Assignment
	persisted map[string]domain.Assignment
	auditErr error
}

func (r *assignmentTransactionRepository) InTx(ctx context.Context, fn func(repository.Repository) error) error {
	// The transaction object shares the parent's persisted store so that direct
	// writes against the parent repository (bypassing the transaction) are
	// detectable, but writes through the transaction land in pending and are
	// only flushed on commit.
	tx := &assignmentTransactionRepository{
		pending:   make(map[string]domain.Assignment),
		persisted: nil, // transaction writes must buffer, not persist directly
		auditErr:  r.auditErr,
	}
	if err := fn(tx); err != nil {
		return err
	}
	for id, a := range tx.pending {
		r.persisted[id] = a
	}
	return nil
}

func (r *assignmentTransactionRepository) CreateAssignment(_ context.Context, assignment domain.Assignment) error {
	if r.persisted != nil {
		// Reaching this branch means CreateAssignment was invoked on the parent
		// repository rather than through the transaction: a write that survives
		// an audit rollback. Fail the test by recording the bypass.
		r.persisted["__direct_write__"] = assignment
		return nil
	}
	r.pending[assignment.ID] = assignment
	return nil
}

func (r *assignmentTransactionRepository) ValidateRolloutCampaignManagedDevice(context.Context, string, string) error {
	return nil
}

func (r *assignmentTransactionRepository) WriteAudit(context.Context, repository.AuditInput) error {
	return r.auditErr
}

func TestAssignManagedDeviceRollsBackWhenAuditFails(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	assignment := domain.Assignment{ID: "a1", RolloutCampaignID: "p", ManagedDeviceID: "s", ReleaseOperatorID: "op", StartsAt: start, EndsAt: start.Add(time.Hour), Status: "queued"}
	operator := domain.ReleaseOperator{ID: "op", Name: "Field", Role: domain.RoleManagedDeviceOperator}
	repo := &assignmentTransactionRepository{
		persisted: map[string]domain.Assignment{},
		auditErr:  errors.New("audit rejected: invalid release_operator identity"),
	}
	err := New(repo).AssignManagedDevice(t.Context(), RequestMeta{RequestID: "req"}, assignment, operator)
	if err == nil {
		t.Fatal("assignment succeeded despite audit failure")
	}
	if len(repo.persisted) != 0 {
		t.Fatalf("persisted assignments=%d, want 0 (audit failure must roll back the assignment)", len(repo.persisted))
	}
}

func TestAssignManagedDeviceCommitsWhenAuditSucceeds(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	assignment := domain.Assignment{ID: "a2", RolloutCampaignID: "p", ManagedDeviceID: "s", ReleaseOperatorID: "op", StartsAt: start, EndsAt: start.Add(time.Hour), Status: "queued"}
	operator := domain.ReleaseOperator{ID: "op", Name: "Field", Role: domain.RoleManagedDeviceOperator}
	repo := &assignmentTransactionRepository{
		persisted: map[string]domain.Assignment{},
		auditErr:  nil,
	}
	if err := New(repo).AssignManagedDevice(t.Context(), RequestMeta{RequestID: "req"}, assignment, operator); err != nil {
		t.Fatalf("assignment error = %v", err)
	}
	if _, ok := repo.persisted["a2"]; !ok {
		t.Fatalf("assignment not persisted on success; persisted=%v", repo.persisted)
	}
}
