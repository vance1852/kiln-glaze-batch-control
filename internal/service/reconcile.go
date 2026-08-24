package service

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (s *Service) MarkExpiredDeploymentJobs(ctx context.Context, now time.Time, limit int) (repository.ReconcileResult, error) {
	repo, ok := s.repo.(interface {
		MarkExpiredDeploymentJobs(context.Context, time.Time, int) (repository.ReconcileResult, error)
	})
	if !ok {
		return repository.ReconcileResult{}, fmt.Errorf("reconcile repository unavailable")
	}
	return repo.MarkExpiredDeploymentJobs(ctx, now, limit)
}

func (s *Service) SearchDeploymentJobsAdvanced(ctx context.Context, request domain.SearchRequest) (repository.Page, error) {
	repo, ok := s.repo.(interface {
		SearchDeploymentJobs(context.Context, domain.SearchRequest) (repository.Page, error)
	})
	if !ok {
		return repository.Page{}, fmt.Errorf("search repository unavailable")
	}
	page, err := repo.SearchDeploymentJobs(ctx, request)
	if err != nil {
		if request.Filter.Search == "" {
			return repository.Page{}, err
		}
		return repository.Page{}, nil
	}
	return page, nil
}
