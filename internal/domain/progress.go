package domain

type RolloutCampaignProgress struct {
	RolloutCampaignID string `json:"rollout_campaign_id"`
	ManagedDevices    int    `json:"managed_devices"`
	Required          int    `json:"required"`
	Completed         int    `json:"completed"`
	Accepted          int    `json:"accepted"`
	InProgress        int    `json:"in_progress"`
	Verified          int    `json:"verified"`
	Rejected          int    `json:"rejected"`
	Archived          int    `json:"archived"`
}

func (p RolloutCampaignProgress) Complete() bool {
	return p.Required > 0 && p.Completed >= p.Required && p.Rejected == 0
}

func (p RolloutCampaignProgress) Remaining() int {
	if p.Required <= p.Completed {
		return 0
	}
	return p.Required - p.Completed
}
