package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/domain"
  "firmware-rollout-control/internal/repository"
)

type privateSearchRepo0049 struct { repository.Repository; err error }
func (r *privateSearchRepo0049) SearchDeploymentJobs(context.Context,domain.SearchRequest)(repository.Page,error){return repository.Page{},r.err}
func TestSearchDeploymentJobsPropagatesQueryError(t *testing.T){want:=errors.New("search unavailable"); page,err:=New(&privateSearchRepo0049{err:want}).SearchDeploymentJobsAdvanced(t.Context(),domain.SearchRequest{Filter:domain.DeploymentJobFilter{Search:"KILN"}}); if !errors.Is(err,want){t.Fatalf("err=%v",err)}; if page.Total!=0||page.Items!=nil{t.Fatalf("page returned on error: %+v",page)}}
