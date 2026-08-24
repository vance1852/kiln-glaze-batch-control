package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRolloutCampaignTransitions(t *testing.T) {
	cases := []struct {
		from, to RolloutCampaignStatus
		want     bool
	}{
		{RolloutCampaignDraft, RolloutCampaignScheduled, true},
		{RolloutCampaignScheduled, RolloutCampaignActive, true},
		{RolloutCampaignActive, RolloutCampaignClosed, true},
		{RolloutCampaignDraft, RolloutCampaignActive, false},
		{RolloutCampaignClosed, RolloutCampaignDraft, false},
	}
	for _, tc := range cases {
		if got := tc.from.CanMoveTo(tc.to); got != tc.want {
			t.Errorf("%s -> %s = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestDeploymentJobMoveSetsTimestampsAndVersion(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	task := DeploymentJob{ID: "s1", TaskCode: "S-1", Status: DeploymentJobQueued, ExpiresAt: now.Add(time.Hour), Version: 3}
	updated, err := task.Move(DeploymentJobCompleted, now)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != DeploymentJobCompleted || updated.Version != 4 || updated.CompletedAt == nil || !updated.CompletedAt.Equal(now) {
		t.Fatalf("unexpected collected task: %+v", updated)
	}
}

func TestDeploymentJobRejectsInvalidMoveAndExpiredDeploymentJob(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	task := DeploymentJob{TaskCode: "S-1", Status: DeploymentJobQueued, ExpiresAt: now.Add(-time.Minute), Version: 1}
	if _, err := task.Move(DeploymentJobAccepted, now); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("invalid transition error = %v", err)
	}
	task.Status = DeploymentJobCompleted
	if _, err := task.Move(DeploymentJobActivationPending, now); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired error = %v", err)
	}
}

func TestInstallationReportOutcomeUsesInclusiveThreshold(t *testing.T) {
	if InstallationReportVerified != InstallationReportStatus("verified") {
		t.Fatal("approved value changed")
	}
	if InstallationReportStatus(InstallationReportVerified).Outcome(10, 10) != InstallationReportVerified {
		t.Fatal("value at limit should be approved")
	}
	if InstallationReportStatus(InstallationReportVerified).Outcome(10.01, 10) != InstallationReportRejected {
		t.Fatal("value over limit should be rejected")
	}
}
