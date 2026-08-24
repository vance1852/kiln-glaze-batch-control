package domain

import (
	"fmt"
	"strings"
)

type ReleaseOperatorRole string

const (
	RoleManagedDeviceOperator ReleaseOperatorRole = "managed_device_operator"
	RoleInstallationOperator  ReleaseOperatorRole = "installation_report_release_operator"
	RoleQualityReviewer       ReleaseOperatorRole = "quality_reviewer"
	RoleSafetySupervisor      ReleaseOperatorRole = "safety_supervisor"
)

type ReleaseOperator struct {
	ID   string              `json:"id"`
	Name string              `json:"name"`
	Role ReleaseOperatorRole `json:"role"`
}

func (o ReleaseOperator) Validate() error {
	if strings.TrimSpace(o.ID) == "" || strings.TrimSpace(o.Name) == "" {
		return fmt.Errorf("release_operator identity is required: %w", ErrConflict)
	}
	switch o.Role {
	case RoleManagedDeviceOperator, RoleInstallationOperator, RoleQualityReviewer, RoleSafetySupervisor:
		return nil
	default:
		return fmt.Errorf("unknown release_operator role: %w", ErrConflict)
	}
}

func (o ReleaseOperator) CanRecordInstallationReport() bool {
	if err := o.Validate(); err != nil {
		return false
	}
	switch o.Role {
	case RoleInstallationOperator, RoleSafetySupervisor:
		return true
	default:
		return false
	}
}

func (o ReleaseOperator) Can(action string) bool {
	switch action {
	case "complete", "activation":
		return o.Role == RoleManagedDeviceOperator || o.Role == RoleSafetySupervisor
	case "record_installation_report":
		return o.Role == RoleInstallationOperator || o.Role == RoleSafetySupervisor
	case "review_installation_report":
		return o.Role == RoleQualityReviewer || o.Role == RoleSafetySupervisor
	case "close_rollout_campaign", "archive":
		return o.Role == RoleSafetySupervisor
	default:
		return false
	}
}
