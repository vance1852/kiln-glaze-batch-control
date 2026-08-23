package service

import (
	"errors"
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
)

func TestCreateDeploymentJobsBulkRejectsEmptyInput(t *testing.T) {
	svc := New(nil)
	if _, err := svc.CreateDeploymentJobsBulk(t.Context(), RequestMeta{}, nil); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestSearchRequestUsesStableDefaults(t *testing.T) {
	request := domain.SearchRequest{}
	request = request.Normalize()
	if request.Sort != domain.SortCreated || request.Limit != 50 || request.Offset != 0 {
		t.Fatalf("request=%+v", request)
	}
}

func TestValidateBulkForManagedDeviceAcceptsSameManagedDevice(t *testing.T) {
	svc := New(nil).WithClock(func() time.Time { return time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC) })
	requests := []domain.DeploymentJobRequest{{RolloutCampaignID: "p", ManagedDeviceID: "s", TaskCode: "S-1", ExpiresAt: time.Date(2026, 8, 19, 9, 0, 0, 0, time.UTC)}}
	if err := svc.ValidateBulkForManagedDevice(requests, "s"); err != nil {
		t.Fatal(err)
	}
}
