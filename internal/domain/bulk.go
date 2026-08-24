package domain

import (
	"fmt"
	"strings"
	"time"
)

type DeploymentJobRequest struct {
	RolloutCampaignID string
	ManagedDeviceID   string
	TaskCode          string
	ExpiresAt         time.Time
}

type BulkItemResult struct {
	Index           int    `json:"index"`
	TaskCode        string `json:"task_code"`
	DeploymentJobID string `json:"deployment_job_id,omitempty"`
	Error           string `json:"error,omitempty"`
}

func (r DeploymentJobRequest) Validate(now time.Time) error {
	if strings.TrimSpace(r.RolloutCampaignID) == "" || strings.TrimSpace(r.ManagedDeviceID) == "" || strings.TrimSpace(r.TaskCode) == "" {
		return fmt.Errorf("task rollout_campaign, managed_device and code are required: %w", ErrConflict)
	}
	if r.ExpiresAt.Before(now) {
		return ErrExpired
	}
	return nil
}

func ValidateBulkRequests(requests []DeploymentJobRequest, now time.Time) error {
	seen := make(map[string]struct{}, len(requests))
	for _, request := range requests {
		if err := request.Validate(now); err != nil {
			return err
		}
		if _, ok := seen[request.TaskCode]; ok {
			return fmt.Errorf("duplicate external code %s: %w", request.TaskCode, ErrConflict)
		}
		seen[request.TaskCode] = struct{}{}
	}
	return nil
}
