package domain

import "fmt"

func CanAssign(release_operator ReleaseOperator, assignment Assignment) error {
	if err := release_operator.Validate(); err != nil {
		return err
	}
	if assignment.ReleaseOperatorID != release_operator.ID {
		return fmt.Errorf("assignment release_operator mismatch: %w", ErrConflict)
	}
	if release_operator.Role != RoleManagedDeviceOperator && release_operator.Role != RoleSafetySupervisor {
		return fmt.Errorf("release_operator cannot receive a managed_device managed_device assignment: %w", ErrConflict)
	}
	return assignment.Validate()
}

func CanReview(release_operator ReleaseOperator, installation_report InstallationReportStatus) error {
	if !release_operator.Has(Perrollout_campaignInstallationReportReview) {
		return fmt.Errorf("release_operator cannot review installation_reports: %w", ErrConflict)
	}
	if installation_report != InstallationReportPending {
		return fmt.Errorf("installation_report is already reviewed: %w", ErrConflict)
	}
	return nil
}
