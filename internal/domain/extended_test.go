package domain

import (
	"errors"
	"testing"
	"time"
)

func TestReleaseOperatorRolesAndPerrollout_campaigns(t *testing.T) {
	cases := []struct {
		role                ReleaseOperatorRole
		perrollout_campaign Perrollout_campaign
		want                bool
	}{
		{RoleManagedDeviceOperator, Perrollout_campaignDeploymentJobComplete, true},
		{RoleManagedDeviceOperator, Perrollout_campaignInstallationReportReview, false},
		{RoleInstallationOperator, Perrollout_campaignInstallationReportRecord, true},
		{RoleQualityReviewer, Perrollout_campaignInstallationReportReview, true},
		{RoleSafetySupervisor, Perrollout_campaignHealthAlertClose, true},
	}
	for _, tc := range cases {
		release_operator := ReleaseOperator{ID: "op", Name: "ReleaseOperator", Role: tc.role}
		if got := release_operator.Has(tc.perrollout_campaign); got != tc.want {
			t.Errorf("role=%s perrollout_campaign=%s got=%v want=%v", tc.role, tc.perrollout_campaign, got, tc.want)
		}
	}
}

func TestReleaseOperatorValidationRejectsUnknownRole(t *testing.T) {
	if err := (ReleaseOperator{ID: "op", Name: "ReleaseOperator", Role: "unknown"}).Validate(); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestAssignmentLifecycle(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	a := Assignment{ID: "a", RolloutCampaignID: "p", ManagedDeviceID: "s", ReleaseOperatorID: "o", StartsAt: start, EndsAt: start.Add(time.Hour), Status: "queued"}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	if !a.CanMoveTo("active") || a.CanMoveTo("completed") {
		t.Fatal("assignment transition graph is wrong")
	}
	if a.ActiveAt(start) {
		t.Fatal("queued assignment is not active")
	}
	a.Status = "active"
	if !a.ActiveAt(start.Add(time.Minute)) {
		t.Fatal("active assignment not active in window")
	}
	if a.ActiveAt(start.Add(time.Hour)) {
		t.Fatal("end boundary should be inactive")
	}
}

func TestDeploymentJobFilterAndSearch(t *testing.T) {
	deployment_jobs := []DeploymentJob{
		{ID: "2", RolloutCampaignID: "p", TaskCode: "S-002", Status: DeploymentJobAccepted, ExpiresAt: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)},
		{ID: "1", RolloutCampaignID: "p", TaskCode: "S-001", Status: DeploymentJobCompleted, ExpiresAt: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)},
		{ID: "3", RolloutCampaignID: "q", TaskCode: "S-003", Status: DeploymentJobAccepted, ExpiresAt: time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)},
	}
	request := SearchRequest{Filter: DeploymentJobFilter{RolloutCampaignID: " p ", Search: "002"}, Sort: SortExpiry, Limit: 10}
	items := SearchDeploymentJobs(deployment_jobs, request)
	if len(items) != 1 || items[0].ID != "2" {
		t.Fatalf("search result = %+v", items)
	}
	if !(DeploymentJobFilter{Status: DeploymentJobAccepted}).Matches(deployment_jobs[0]) {
		t.Fatal("status filter did not match")
	}
}

func TestBulkValidationDetectsDuplicateAndExpiredInput(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	valid := []DeploymentJobRequest{{RolloutCampaignID: "p", ManagedDeviceID: "s", TaskCode: "S-1", ExpiresAt: now.Add(time.Hour)}}
	if err := ValidateBulkRequests(valid, now); err != nil {
		t.Fatal(err)
	}
	duplicate := append(valid, valid[0])
	if err := ValidateBulkRequests(duplicate, now); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate error = %v", err)
	}
	expired := []DeploymentJobRequest{{RolloutCampaignID: "p", ManagedDeviceID: "s", TaskCode: "S-2", ExpiresAt: now.Add(-time.Second)}}
	if err := ValidateBulkRequests(expired, now); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired error = %v", err)
	}
}

func TestStateMachinesReachExpectedStates(t *testing.T) {
	rollout_campaignStates := DefaultRolloutCampaignMachine().Reachable("draft")
	if len(rollout_campaignStates) != 4 {
		t.Fatalf("rollout_campaign states = %v", rollout_campaignStates)
	}
	if err := DefaultDeploymentJobMachine().ValidatePath([]string{"queued", "completed", "activation_pending", "accepted", "in_progress", "rejected", "archived"}); err != nil {
		t.Fatal(err)
	}
	if err := DefaultRolloutWaveMachine().ValidatePath([]string{"queued", "completed"}); err == nil {
		t.Fatal("invalid rollout_wave path accepted")
	}
}

func TestConstraintSetAndRedaction(t *testing.T) {
	if err := (ConstraintSet{MaxDeploymentJobsPerRolloutWave: 2, MinimumRemainingTTL: time.Hour}).Validate(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	task := DeploymentJob{Status: DeploymentJobAccepted, ExpiresAt: now.Add(2 * time.Hour)}
	if !(ConstraintSet{MaxDeploymentJobsPerRolloutWave: 2, MinimumRemainingTTL: time.Hour}).AllowsDeploymentJob(task, now) {
		t.Fatal("valid task rejected")
	}
	if RedactTaskCode("S-1234") != "S-**34" {
		t.Fatalf("redaction mismatch")
	}
	if RedactRolloutLane("North Gate") != "N***e" {
		t.Fatalf("rollout_lane redaction mismatch")
	}
}

func TestValidationHelpers(t *testing.T) {
	if err := ValidateBusinessCode("PLAN-001"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBusinessCode("bad code"); !errors.Is(err, ErrConflict) {
		t.Fatalf("code error = %v", err)
	}
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	if err := ValidateUTCWindow(start, start.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePositiveVersion(0); !errors.Is(err, ErrConflict) {
		t.Fatalf("version error = %v", err)
	}
	if err := ValidatePage(0, 101); !errors.Is(err, ErrConflict) {
		t.Fatalf("page error = %v", err)
	}
	if err := ValidateInstallationReport(1, 2, "mg/L"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReason("ok"); !errors.Is(err, ErrConflict) {
		t.Fatalf("reason error = %v", err)
	}
}
