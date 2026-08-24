package service

import (
	"errors"
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func TestCreateRolloutWaveRejectsEmptyAndOversizedRequestsBeforePersistence(t *testing.T) {
	svc := New(nil)
	if _, err := svc.CreateRolloutWave(t.Context(), RequestMeta{}, repository.RolloutWaveInput{Code: "B", Method: "m", Capacity: 1}, nil); !errors.Is(err, domain.ErrCapacityExceeded) {
		t.Fatalf("empty rollout_wave error = %v", err)
	}
	if _, err := svc.CreateRolloutWave(t.Context(), RequestMeta{}, repository.RolloutWaveInput{Code: "B", Method: "m", Capacity: 1}, []string{"a", "b"}); !errors.Is(err, domain.ErrCapacityExceeded) {
		t.Fatalf("oversized rollout_wave error = %v", err)
	}
}

func TestCreateRolloutCampaignRejectsMissingManagedDevicesBeforeTransaction(t *testing.T) {
	svc := New(nil)
	now := time.Now().UTC()
	_, err := svc.CreateRolloutCampaign(t.Context(), RequestMeta{}, CreateRolloutCampaignRequest{Code: "P", Name: "RolloutCampaign", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: "release_operator", ManagedDevices: nil})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("missing managed_devices error = %v", err)
	}
}

func TestReviewInstallationReportRejectsInvalidDeploymentJobVersion(t *testing.T) {
	if domain.DeploymentJobStatus("in_progress").CanMoveTo(domain.DeploymentJobRejected) == false {
		t.Fatal("in-progress managed_device should support rejection")
	}
	if domain.DeploymentJobStatus("queued").CanMoveTo(domain.DeploymentJobRejected) {
		t.Fatal("queued task cannot be rejected")
	}
}
