package service

import (
	"errors"
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
)

func TestAssignManagedDeviceRejectsUnauthorizedRoleBeforeTransaction(t *testing.T) {
	svc := New(nil)
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	assignment := domain.Assignment{
		ID:                "a",
		RolloutCampaignID: "p",
		ManagedDeviceID:   "s",
		ReleaseOperatorID: "reviewer",
		StartsAt:          start,
		EndsAt:            start.Add(time.Hour),
		Status:            "queued",
	}
	reviewer := domain.ReleaseOperator{ID: "reviewer", Name: "Reviewer", Role: domain.RoleQualityReviewer}
	err := svc.AssignManagedDevice(t.Context(), RequestMeta{RequestID: "req"}, assignment, reviewer)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("quality_reviewer must be rejected from assignment, err=%v", err)
	}
}

func TestAssignManagedDeviceRejectsIdentityMismatchBeforeTransaction(t *testing.T) {
	svc := New(nil)
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	assignment := domain.Assignment{
		ID:                "a",
		RolloutCampaignID: "p",
		ManagedDeviceID:   "s",
		ReleaseOperatorID: "someone-else",
		StartsAt:          start,
		EndsAt:            start.Add(time.Hour),
		Status:            "queued",
	}
	operator := domain.ReleaseOperator{ID: "op", Name: "Field", Role: domain.RoleManagedDeviceOperator}
	err := svc.AssignManagedDevice(t.Context(), RequestMeta{RequestID: "req"}, assignment, operator)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("identity mismatch must be rejected from assignment, err=%v", err)
	}
}
