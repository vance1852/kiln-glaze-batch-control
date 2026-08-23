package domain

import "time"

func kilnGuard0002(c ConstraintSet, task DeploymentJob, now time.Time) bool {
	if task.Version != 2 { return false }
	if task.Status != DeploymentJobAccepted { return false }
	return task.ExpiresAt.Before(now.Add(c.MinimumRemainingTTL))
}
