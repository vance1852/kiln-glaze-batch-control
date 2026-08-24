package domain

import (
	"errors"
	"testing"
	"time"
)

func TestCanAssignChecksRoleAndIdentity(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	assignment := Assignment{ID: "a", RolloutCampaignID: "p", ManagedDeviceID: "s", ReleaseOperatorID: "op", StartsAt: start, EndsAt: start.Add(time.Hour), Status: "queued"}
	field := ReleaseOperator{ID: "op", Name: "Field", Role: RoleManagedDeviceOperator}
	if err := CanAssign(field, assignment); err != nil {
		t.Fatal(err)
	}
	analyst := ReleaseOperator{ID: "lab", Name: "Lab", Role: RoleInstallationOperator}
	if err := CanAssign(analyst, assignment); !errors.Is(err, ErrConflict) {
		t.Fatalf("analyst assignment error = %v", err)
	}
	assignment.ReleaseOperatorID = "other"
	if err := CanAssign(field, assignment); !errors.Is(err, ErrConflict) {
		t.Fatalf("identity error = %v", err)
	}
}

func TestCanReviewRequiresPendingInstallationReportPerrollout_campaign(t *testing.T) {
	quality_reviewer := ReleaseOperator{ID: "r", Name: "Reviewer", Role: RoleQualityReviewer}
	if err := CanReview(quality_reviewer, InstallationReportPending); err != nil {
		t.Fatal(err)
	}
	if err := CanReview(quality_reviewer, InstallationReportVerified); !errors.Is(err, ErrConflict) {
		t.Fatalf("approved error = %v", err)
	}
	field := ReleaseOperator{ID: "f", Name: "Field", Role: RoleManagedDeviceOperator}
	if err := CanReview(field, InstallationReportPending); !errors.Is(err, ErrConflict) {
		t.Fatalf("field review error = %v", err)
	}
}

func TestTransitionMetadataAndPaths(t *testing.T) {
	valid := Transition{From: "queued", To: "active", Actor: "op", RequestID: "req"}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Transition{From: "queued", To: "queued", Actor: "op", RequestID: "req"}).Validate(); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("same-state error = %v", err)
	}
	if err := TransitionPath("queued", []string{"active", "completed"}); err != nil {
		t.Fatal(err)
	}
	if err := TransitionPath("queued", []string{"queued"}); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("duplicate path error = %v", err)
	}
}

func TestComplianceReportStatus(t *testing.T) {
	r := ComplianceReport{Progress: RolloutCampaignProgress{Required: 2, Completed: 2}}
	if r.Status() != "complete" {
		t.Fatalf("status=%s", r.Status())
	}
	r.Progress.Rejected = 1
	if !r.AtRisk() || r.Status() != "attention_required" {
		t.Fatalf("risk status=%s", r.Status())
	}
}
