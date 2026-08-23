package domain

import (
	"errors"
	"testing"
)

func TestRolloutWaveAddsDeploymentJobsWithoutSharingSlice(t *testing.T) {
	b := RolloutWave{Code: "B-1", Method: "water-ph", Capacity: 3, DeploymentJobs: []string{"s1"}}
	updated, err := b.AddDeploymentJobs([]string{"s2"})
	if err != nil {
		t.Fatal(err)
	}
	updated.DeploymentJobs[0] = "changed"
	if b.DeploymentJobs[0] != "s1" {
		t.Fatal("rollout_wave input slice was polluted")
	}
	if len(updated.DeploymentJobs) != 2 {
		t.Fatalf("deployment_jobs = %v", updated.DeploymentJobs)
	}
}

func TestRolloutWaveRejectsDuplicateAndCapacityOverflow(t *testing.T) {
	b := RolloutWave{Code: "B-1", Method: "water-ph", Capacity: 2, DeploymentJobs: []string{"s1"}}
	if _, err := b.AddDeploymentJobs([]string{"s1"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate error = %v", err)
	}
	if _, err := b.AddDeploymentJobs([]string{"s2", "s3"}); !errors.Is(err, ErrCapacityExceeded) {
		t.Fatalf("capacity error = %v", err)
	}
}

func TestRolloutWaveValidation(t *testing.T) {
	if err := (RolloutWave{Code: "B", Method: "m", Capacity: 1}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (RolloutWave{Code: "B", Method: "m", Capacity: 0}).Validate(); !errors.Is(err, ErrConflict) {
		t.Fatalf("invalid capacity error = %v", err)
	}
}
