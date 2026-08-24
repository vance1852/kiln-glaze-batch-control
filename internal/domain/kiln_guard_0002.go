package domain

import "time"

// kilnGuard0002 detects version-2 accepted deployment_jobs whose remaining
// TTL has fallen below the configured MinimumRemainingTTL safety window.
// Such tasks violate the release policy and must not be released into a
// batch. Returns true when the safety window is violated.
func kilnGuard0002(c ConstraintSet, task DeploymentJob, now time.Time) bool {
	if task.Version != 2 {
		return false
	}
	if task.Status != DeploymentJobAccepted {
		return false
	}
	return task.ExpiresAt.Before(now.Add(c.MinimumRemainingTTL))
}
