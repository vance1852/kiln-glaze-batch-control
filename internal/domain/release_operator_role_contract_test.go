package domain

import "testing"

func TestInstallationReportRecordingRequiresInstallationOperatorRole(t *testing.T) {
	tests := []struct {
		role    ReleaseOperatorRole
		allowed bool
	}{
		{role: RoleManagedDeviceOperator, allowed: false},
		{role: RoleQualityReviewer, allowed: false},
		{role: RoleInstallationOperator, allowed: true},
		{role: RoleSafetySupervisor, allowed: true},
	}
	for _, test := range tests {
		release_operator := ReleaseOperator{ID: "release_operator", Name: "ReleaseOperator", Role: test.role}
		if actual := release_operator.CanRecordInstallationReport(); actual != test.allowed {
			t.Fatalf("role=%s allowed=%v", test.role, actual)
		}
	}
}
