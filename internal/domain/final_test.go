package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSearchDescendingSortAndPagination(t *testing.T) {
	items := []DeploymentJob{{ID: "a", TaskCode: "A", Status: DeploymentJobAccepted}, {ID: "c", TaskCode: "C", Status: DeploymentJobAccepted}, {ID: "b", TaskCode: "B", Status: DeploymentJobAccepted}}
	got := SearchDeploymentJobs(items, SearchRequest{Sort: SortCode, Desc: true, Limit: 2})
	if len(got) != 2 || got[0].ID != "c" || got[1].ID != "b" {
		t.Fatalf("got=%+v", got)
	}
}

func TestSameRolloutCampaignRequiresAllDeploymentJobsToMatch(t *testing.T) {
	if !SameRolloutCampaign([]DeploymentJob{{RolloutCampaignID: "p"}, {RolloutCampaignID: "p"}}) {
		t.Fatal("same cases rejected")
	}
	if SameRolloutCampaign([]DeploymentJob{{RolloutCampaignID: "p"}, {RolloutCampaignID: "q"}}) {
		t.Fatal("different cases accepted")
	}
}

func TestHealthAlertRejectsTooOldDueDate(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	d := HealthAlert{DeploymentJobID: "s", Kind: "reassess", Status: HealthAlertOpen, Reason: "bad", DueAt: now.Add(-25 * time.Hour)}
	if err := d.Validate(now); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestRolloutWaveRejectsBlankDeploymentJobID(t *testing.T) {
	b := RolloutWave{Code: "B", Method: "m", Capacity: 2}
	if _, err := b.AddDeploymentJobs([]string{""}); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestUTCWindowRejectsLocalTime(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)
	end := start.Add(time.Hour)
	if err := ValidateUTCWindow(start, end); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestRolloutCampaignSummaryContainsStableFields(t *testing.T) {
	p := RolloutCampaign{ID: "p", Code: "P-1", Status: RolloutCampaignDraft, Version: 3}
	summary := p.Summary()
	if summary["id"] != "p" || summary["version"] != int64(3) {
		t.Fatalf("summary=%+v", summary)
	}
}
