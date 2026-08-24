package service

import (
	"context"
	"errors"
	"testing"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

type release_operatorTransactionRepository struct {
	repository.Repository
	release_operators map[string]domain.ReleaseOperator
	auditErr          error
}

func (r *release_operatorTransactionRepository) InTx(ctx context.Context, fn func(repository.Repository) error) error {
	pending := make(map[string]domain.ReleaseOperator, len(r.release_operators))
	for id, release_operator := range r.release_operators {
		pending[id] = release_operator
	}
	tx := &release_operatorTransactionRepository{release_operators: pending, auditErr: r.auditErr}
	if err := fn(tx); err != nil {
		return err
	}
	r.release_operators = pending
	return nil
}

func (r *release_operatorTransactionRepository) CreateReleaseOperatorOutsideTransaction(_ context.Context, release_operator domain.ReleaseOperator) error {
	r.release_operators[release_operator.ID] = release_operator
	return nil
}

func (r *release_operatorTransactionRepository) CreateReleaseOperator(_ context.Context, release_operator domain.ReleaseOperator) error {
	r.release_operators[release_operator.ID] = release_operator
	return nil
}

func (r *release_operatorTransactionRepository) WriteAudit(context.Context, repository.AuditInput) error {
	return r.auditErr
}

func TestReleaseOperatorRegistrationRollsBackWhenAuditFails(t *testing.T) {
	repo := &release_operatorTransactionRepository{release_operators: map[string]domain.ReleaseOperator{}, auditErr: errors.New("audit rejected")}
	_, err := New(repo).RegisterReleaseOperator(t.Context(), RequestMeta{RequestID: "release_operator-create"}, "Rollback ReleaseOperator", domain.RoleSafetySupervisor)
	if err == nil {
		t.Fatal("release_operator registration succeeded despite audit failure")
	}
	if len(repo.release_operators) != 0 {
		t.Fatalf("persisted release_operators=%d", len(repo.release_operators))
	}
}
