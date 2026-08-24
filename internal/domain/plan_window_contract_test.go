package domain

import (
	"testing"
	"time"
)

func TestCollectionRequiresAnOpenRolloutCampaignWindow(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	rollout_campaign := RolloutCampaign{Status: RolloutCampaignActive, StartsAt: now.Add(-2 * time.Hour), EndsAt: now.Add(-time.Hour)}
	if RolloutCampaignExecutionAllowed(rollout_campaign, now) {
		t.Fatal("collection allowed after rollout_campaign window ended")
	}
}
