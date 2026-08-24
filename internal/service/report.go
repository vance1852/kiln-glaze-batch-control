package service

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
)

func (s *Service) ComplianceReport(ctx context.Context, rollout_campaignID string) (domain.ComplianceReport, error) {
	if _, err := s.repo.GetRolloutCampaign(ctx, rollout_campaignID); err != nil {
		return domain.ComplianceReport{}, err
	}
	repo, ok := s.repo.(interface {
		ComplianceReport(context.Context, string, time.Time) (domain.ComplianceReport, error)
	})
	if !ok {
		return domain.ComplianceReport{}, fmt.Errorf("report repository unavailable")
	}
	return repo.ComplianceReport(ctx, rollout_campaignID, s.now())
}

func (s *Service) PublicDeploymentJob(task domain.DeploymentJob) map[string]any {
	return map[string]any{"id": task.ID, "task_code": domain.RedactTaskCode(task.TaskCode), "status": task.Status, "expires_at": task.ExpiresAt, "version": task.Version}
}
