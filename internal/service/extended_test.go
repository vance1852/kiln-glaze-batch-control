package service

import (
	"errors"
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func TestAuthorizeRequiresRepositoryAndPerrollout_campaign(t *testing.T) {
	svc := New(nil)
	if err := svc.Authorize(t.Context(), "release_operator", "complete"); err == nil {
		t.Fatal("authorization succeeded without repository")
	}
}

func TestOpenHealthAlertRejectsOldDueTime(t *testing.T) {
	svc := New(nil).WithClock(func() time.Time { return time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC) })
	_, err := svc.OpenHealthAlert(t.Context(), RequestMeta{}, repository.HealthAlertInput{DeploymentJobID: "s1", Kind: "reassess", Reason: "bad", DueAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestValidateBulkForManagedDeviceRejectsMixedManagedDevices(t *testing.T) {
	svc := New(nil).WithClock(time.Now)
	requests := []domain.DeploymentJobRequest{{RolloutCampaignID: "p", ManagedDeviceID: "s1", TaskCode: "S-1", ExpiresAt: time.Now().Add(time.Hour)}, {RolloutCampaignID: "p", ManagedDeviceID: "s2", TaskCode: "S-2", ExpiresAt: time.Now().Add(time.Hour)}}
	if err := svc.ValidateBulkForManagedDevice(requests, "s1"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}
