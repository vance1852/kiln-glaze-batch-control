package domain

import "time"

type ComplianceReport struct {
	RolloutCampaignID string
	GeneratedAt       time.Time
	Progress          RolloutCampaignProgress
	Expiring          []DeploymentJob
	OpenHealthAlerts  int
}

func (r ComplianceReport) AtRisk() bool {
	return len(r.Expiring) > 0 || r.OpenHealthAlerts > 0 || r.Progress.Rejected > 0
}

func (r ComplianceReport) Status() string {
	if r.AtRisk() {
		return "attention_required"
	}
	if r.Progress.Complete() {
		return "complete"
	}
	return "in_progress"
}
