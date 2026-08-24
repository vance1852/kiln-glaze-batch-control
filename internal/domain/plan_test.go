package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRolloutCampaignWindowValidation(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	valid := RolloutCampaign{Timezone: "Asia/Shanghai", StartsAt: now, EndsAt: now.Add(time.Hour), Status: RolloutCampaignDraft}
	if err := valid.ValidateWindow(now); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.EndsAt = invalid.StartsAt
	if err := invalid.ValidateWindow(now); !errors.Is(err, ErrConflict) {
		t.Fatalf("equal window error = %v", err)
	}
	late := valid
	late.Status = RolloutCampaignScheduled
	if err := late.ValidateWindow(now.Add(2 * time.Hour)); !errors.Is(err, ErrExpired) {
		t.Fatalf("late window error = %v", err)
	}
}

func TestRolloutCampaignCollectionWindow(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	p := RolloutCampaign{Status: RolloutCampaignActive, StartsAt: start, EndsAt: start.Add(time.Hour)}
	if !p.CanExecuteAt(start) {
		t.Fatal("start boundary should be included")
	}
	if p.CanExecuteAt(start.Add(time.Hour)) {
		t.Fatal("end boundary should be excluded")
	}
	if p.RemainingWindow(start.Add(30*time.Minute)) != 30*time.Minute {
		t.Fatal("remaining window mismatch")
	}
	if p.RemainingWindow(start.Add(2*time.Hour)) != 0 {
		t.Fatal("expired window should be zero")
	}
}

func TestManagedDeviceValidation(t *testing.T) {
	managed_device := ManagedDevice{Code: "S-1", RolloutLane: "A-101", RequiredSuccesses: 2}
	if err := managed_device.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []ManagedDevice{{RolloutLane: "x", RequiredSuccesses: 1}, {Code: "x", RequiredSuccesses: 0}, {Code: "x", RolloutLane: "x", RequiredSuccesses: 2, Completed: 3}} {
		if err := invalid.Validate(); !errors.Is(err, ErrConflict) {
			t.Fatalf("invalid managed_device error = %v", err)
		}
	}
}
