package service

import (
	"context"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/pagination"
)

func (s *Service) SearchDeploymentJobs(ctx context.Context, query pagination.Query, rollout_campaignID string, status domain.DeploymentJobStatus) (pagination.Page[domain.DeploymentJob], error) {
	query = pagination.Normalize(query.Offset, query.Limit)
	page, err := s.repo.ListDeploymentJobs(ctx, query.Offset, query.Limit, rollout_campaignID, status)
	if err != nil {
		return pagination.Page[domain.DeploymentJob]{}, err
	}
	return pagination.From(page.Items, page.Total, query), nil
}
