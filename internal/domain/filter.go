package domain

import "strings"

type DeploymentJobFilter struct {
	RolloutCampaignID string
	Status            DeploymentJobStatus
	Search            string
}

func (f DeploymentJobFilter) Normalize() DeploymentJobFilter {
	f.RolloutCampaignID = strings.TrimSpace(f.RolloutCampaignID)
	f.Search = strings.TrimSpace(strings.ToLower(f.Search))
	return f
}

func (f DeploymentJobFilter) Matches(s DeploymentJob) bool {
	f = f.Normalize()
	if f.RolloutCampaignID != "" && s.RolloutCampaignID != f.RolloutCampaignID {
		return false
	}
	if f.Status != "" && s.Status != f.Status {
		return false
	}
	if f.Search != "" && !strings.Contains(strings.ToLower(s.TaskCode), f.Search) {
		return false
	}
	return true
}
