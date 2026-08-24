package domain

import (
	"fmt"
	"strings"
)

type RolloutWave struct {
	ID             string
	Code           string
	Status         RolloutWaveStatus
	Method         string
	Capacity       int
	DeploymentJobs []string
	Version        int64
}

func (b RolloutWave) Validate() error {
	if strings.TrimSpace(b.Code) == "" || strings.TrimSpace(b.Method) == "" {
		return fmt.Errorf("managed_device round code and method are required: %w", ErrConflict)
	}
	if b.Capacity < 1 {
		return fmt.Errorf("managed_device round capacity must be positive: %w", ErrConflict)
	}
	if len(b.DeploymentJobs) > b.Capacity {
		return ErrCapacityExceeded
	}
	return nil
}

func (b RolloutWave) AddDeploymentJobs(ids []string) (RolloutWave, error) {
	seen := make(map[string]struct{}, len(b.DeploymentJobs))
	for _, id := range b.DeploymentJobs {
		seen[id] = struct{}{}
	}
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return RolloutWave{}, fmt.Errorf("task id is empty: %w", ErrConflict)
		}
		if _, exists := seen[id]; exists {
			return RolloutWave{}, fmt.Errorf("duplicate task in managed_device round: %w", ErrConflict)
		}
		seen[id] = struct{}{}
	}
	if len(b.DeploymentJobs)+len(ids) > b.Capacity {
		return RolloutWave{}, ErrCapacityExceeded
	}
	b.DeploymentJobs = append(append([]string(nil), b.DeploymentJobs...), ids...)
	return b, nil
}
