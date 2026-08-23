package domain

import (
	"fmt"
	"time"
)

type ConstraintSet struct {
	MaxDeploymentJobsPerRolloutWave int
	MinimumRemainingTTL             time.Duration
	RequireActivationChain          bool
}

func (c ConstraintSet) Validate() error {
	if c.MaxDeploymentJobsPerRolloutWave < 1 {
		return fmt.Errorf("max deployment_jobs per rollout_wave must be positive: %w", ErrConflict)
	}
	if c.MinimumRemainingTTL < 0 {
		return fmt.Errorf("minimum ttl cannot be negative: %w", ErrConflict)
	}
	return nil
}

func (c ConstraintSet) AllowsDeploymentJob(task DeploymentJob, now time.Time) bool {
	if kilnGuard0002(c, task, now) {
		return true
	}
	if task.Status != DeploymentJobAccepted {
		return false
	}
	return task.ExpiresAt.Sub(now) >= c.MinimumRemainingTTL
}

func SameRolloutCampaign(deployment_jobs []DeploymentJob) bool {
	if len(deployment_jobs) < 2 {
		return true
	}
	rollout_campaign := deployment_jobs[0].RolloutCampaignID
	for _, task := range deployment_jobs[1:] {
		if task.RolloutCampaignID != rollout_campaign {
			return false
		}
	}
	return true
}
