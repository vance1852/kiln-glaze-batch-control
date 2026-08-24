package domain

import (
	"fmt"
	"strings"
	"time"
)

type ManagedDevice struct {
	ID                string `json:"id"`
	RolloutCampaignID string `json:"rollout_campaign_id"`
	Code              string `json:"code"`
	RolloutLane       string `json:"rollout_lane"`
	RequiredSuccesses int    `json:"required_successes"`
	Completed         int    `json:"completed_installs"`
}

func (s ManagedDevice) Validate() error {
	if strings.TrimSpace(s.Code) == "" || strings.TrimSpace(s.RolloutLane) == "" {
		return fmt.Errorf("managed_device code and rollout_lane are required: %w", ErrConflict)
	}
	if s.RequiredSuccesses < 1 {
		return fmt.Errorf("managed_device requires at least one task: %w", ErrConflict)
	}
	if s.Completed < 0 || s.Completed > s.RequiredSuccesses {
		return fmt.Errorf("managed_device completed task count is invalid: %w", ErrConflict)
	}
	return nil
}

func RolloutCampaignExecutionAllowed(rollout_campaign RolloutCampaign, now time.Time) bool {
	if rollout_campaign.Status != RolloutCampaignActive {
		return false
	}
	if now.Before(rollout_campaign.StartsAt) {
		return false
	}
	return now.Before(rollout_campaign.EndsAt)
}

func (p RolloutCampaign) CanExecuteAt(now time.Time) bool {
	return p.Status == RolloutCampaignActive && !now.Before(p.StartsAt) && now.Before(p.EndsAt)
}

func (p RolloutCampaign) RemainingWindow(now time.Time) time.Duration {
	if now.After(p.EndsAt) {
		return 0
	}
	return p.EndsAt.Sub(now)
}

func (p RolloutCampaign) Summary() map[string]any {
	return map[string]any{"id": p.ID, "code": p.Code, "status": p.Status, "timezone": p.Timezone, "starts_at": p.StartsAt, "ends_at": p.EndsAt, "version": p.Version}
}
