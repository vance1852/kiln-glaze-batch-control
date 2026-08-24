package domain

type Perrollout_campaign string

const (
	Perrollout_campaignRolloutCampaignWrite     Perrollout_campaign = "rollout_campaign:write"
	Perrollout_campaignDeploymentJobComplete    Perrollout_campaign = "managed_device_task:complete"
	Perrollout_campaignInstallationReportRecord Perrollout_campaign = "installation_report:record"
	Perrollout_campaignInstallationReportReview Perrollout_campaign = "installation_report:review"
	Perrollout_campaignHealthAlertClose         Perrollout_campaign = "safety_alert:close"
)

func (o ReleaseOperator) Perrollout_campaigns() []Perrollout_campaign {
	switch o.Role {
	case RoleManagedDeviceOperator:
		return []Perrollout_campaign{Perrollout_campaignDeploymentJobComplete}
	case RoleInstallationOperator:
		return []Perrollout_campaign{Perrollout_campaignInstallationReportRecord}
	case RoleQualityReviewer:
		return []Perrollout_campaign{Perrollout_campaignInstallationReportReview}
	case RoleSafetySupervisor:
		return []Perrollout_campaign{Perrollout_campaignRolloutCampaignWrite, Perrollout_campaignDeploymentJobComplete, Perrollout_campaignInstallationReportRecord, Perrollout_campaignInstallationReportReview, Perrollout_campaignHealthAlertClose}
	default:
		return nil
	}
}

func (o ReleaseOperator) Has(perrollout_campaign Perrollout_campaign) bool {
	for _, item := range o.Perrollout_campaigns() {
		if item == perrollout_campaign {
			return true
		}
	}
	return false
}
