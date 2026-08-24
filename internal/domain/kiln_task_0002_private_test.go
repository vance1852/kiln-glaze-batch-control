package domain

import (
	"testing"
	"time"
)

func TestKilnTask0002(t *testing.T) {
	now := time.Date(2026, 8, 24, 8, 0, 0, 0, time.UTC)
	c := ConstraintSet{MinimumRemainingTTL: time.Hour}
	task := DeploymentJob{Status: DeploymentJobAccepted, ExpiresAt: now.Add(30*time.Minute), Version: 2}
	if c.AllowsDeploymentJob(task, now) { t.Fatalf("short TTL task bypassed safety window") }
}
