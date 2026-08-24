package domain

import (
	"testing"
	"time"
)

func TestInProgressEligibilityRequiresCurrentAcceptedDeploymentJob(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		status    DeploymentJobStatus
		expiresAt time.Time
	}{
		{name: "not accepted", status: DeploymentJobQueued, expiresAt: now.Add(time.Hour)},
		{name: "expired", status: DeploymentJobAccepted, expiresAt: now.Add(-time.Minute)},
	}
	for _, test := range tests {
		if EligibleForExecution(test.status, test.expiresAt, now) {
			t.Fatalf("%s task is eligible", test.name)
		}
	}
}
