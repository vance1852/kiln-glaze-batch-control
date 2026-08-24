package rollout

import (
	"fmt"
	"slices"
	"strings"
)

func IdempotencyKey(tenantID, method, path, key string) (string, error) {
	values := []string{tenantID, method, path, key}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("idempotency scope is incomplete: %w", ErrInvalid)
		}
	}
	return strings.Join(values, "\x00"), nil
}

func CallbackKey(callback Callback) (string, error) {
	if callback.TenantID == "" || callback.DeviceID == "" || callback.ArtifactID == "" || callback.EventID == "" {
		return "", fmt.Errorf("callback identity is incomplete: %w", ErrInvalid)
	}
	return strings.Join([]string{callback.TenantID, callback.DeviceID, callback.ArtifactID, callback.EventID}, "\x00"), nil
}

func ArtifactSupports(artifact Artifact, device Device) error {
	if artifact.TenantID != device.TenantID {
		return fmt.Errorf("artifact and device tenants differ: %w", ErrConflict)
	}
	if !artifact.Signed || artifact.Digest == "" {
		return fmt.Errorf("artifact signature is not verified: %w", ErrConflict)
	}
	if !slices.Contains(artifact.DeviceClasses, device.Class) {
		return fmt.Errorf("artifact does not support device class %s: %w", device.Class, ErrConflict)
	}
	return nil
}

func CanDispatch(campaign Campaign) error {
	if campaign.State != "running" {
		return fmt.Errorf("campaign %s cannot dispatch from %s: %w", campaign.ID, campaign.State, ErrConflict)
	}
	if campaign.ApprovedDigest == "" {
		return fmt.Errorf("campaign %s has no approved artifact digest: %w", campaign.ID, ErrConflict)
	}
	return nil
}

func CanPromote(campaign Campaign) error {
	if campaign.State != "running" {
		return fmt.Errorf("campaign %s is not running: %w", campaign.ID, ErrConflict)
	}
	if campaign.Failed > 0 || campaign.Healthy < campaign.RequiredHealthy {
		return fmt.Errorf("canary health gate is not satisfied: %w", ErrConflict)
	}
	return nil
}

func RollbackVersion(device Device) (string, error) {
	if device.PreviousVersion == "" || device.PreviousVersion == device.CurrentVersion {
		return "", fmt.Errorf("device %s has no rollback target: %w", device.ID, ErrConflict)
	}
	return device.PreviousVersion, nil
}

func CheckGeneration(device Device, expected int64) error {
	if expected <= 0 || device.Generation != expected {
		return fmt.Errorf("device generation changed: %w", ErrConflict)
	}
	return nil
}

func CanCloseAlert(campaign Campaign, openAlerts int) error {
	if openAlerts < 0 {
		return fmt.Errorf("open alert count is invalid: %w", ErrInvalid)
	}
	if openAlerts > 0 || campaign.State == "rolling_back" {
		return fmt.Errorf("campaign still has unresolved safety work: %w", ErrConflict)
	}
	return nil
}
