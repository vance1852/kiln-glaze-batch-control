package domain

import "testing"

func TestStatusCountsIncludesOnlyObservedStatuses(t *testing.T) {
	counts := StatusCounts([]DeploymentJob{{Status: DeploymentJobAccepted}, {Status: DeploymentJobAccepted}, {Status: DeploymentJobRejected}})
	if counts[DeploymentJobAccepted] != 2 || counts[DeploymentJobRejected] != 1 || counts[DeploymentJobVerified] != 0 {
		t.Fatalf("counts=%v", counts)
	}
}

func TestRolloutCampaignProgressCompleteRequiresNoRejectedDeploymentJobs(t *testing.T) {
	progress := RolloutCampaignProgress{Required: 2, Completed: 2}
	if !progress.Complete() {
		t.Fatal("complete progress rejected")
	}
	progress.Rejected = 1
	if progress.Complete() {
		t.Fatal("rejected progress marked complete")
	}
}

func TestReleaseOperatorCanAction(t *testing.T) {
	release_operator := ReleaseOperator{ID: "o", Name: "Supervisor", Role: RoleSafetySupervisor}
	if !release_operator.Can("archive") || !release_operator.Can("review_installation_report") {
		t.Fatal("safety_supervisor perrollout_campaigns missing")
	}
	field := ReleaseOperator{ID: "f", Name: "Field", Role: RoleManagedDeviceOperator}
	if field.Can("archive") {
		t.Fatal("managed_device_operator can archive")
	}
}
