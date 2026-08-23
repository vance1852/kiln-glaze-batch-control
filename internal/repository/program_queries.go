package repository

import (
	"context"
	"fmt"

	"firmware-rollout-control/internal/domain"
)

type RolloutCampaignFilter struct {
	Status domain.RolloutCampaignStatus
	Search string
	Limit  int
	Offset int
}

func (p *Postgres) ListRolloutCampaigns(ctx context.Context, filter RolloutCampaignFilter) ([]domain.RolloutCampaign, int, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	where := "WHERE TRUE"
	args := []any{filter.Limit, filter.Offset}
	if filter.Status != "" {
		args = append(args, filter.Status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		where += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", len(args), len(args))
	}
	rows, err := p.pool.Query(ctx, fmt.Sprintf(`SELECT id,code,name,status,timezone,starts_at,ends_at,version,created_by FROM rollout_campaigns %s ORDER BY starts_at DESC LIMIT $1 OFFSET $2`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list cases: %w", err)
	}
	defer rows.Close()
	items := make([]domain.RolloutCampaign, 0)
	for rows.Next() {
		var item domain.RolloutCampaign
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Status, &item.Timezone, &item.StartsAt, &item.EndsAt, &item.Version, &item.CreatedBy); err != nil {
			return nil, 0, fmt.Errorf("scan rollout_campaign: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	countArgs := args[2:]
	countWhere := "WHERE TRUE"
	if filter.Status != "" {
		countWhere += " AND status=$1"
	}
	if filter.Search != "" {
		index := 1
		if filter.Status != "" {
			index = 2
		}
		countWhere += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", index, index)
	}
	var total int
	if err := p.pool.QueryRow(ctx, "SELECT count(*) FROM rollout_campaigns "+countWhere, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count cases: %w", err)
	}
	return items, total, nil
}

func (p *Postgres) ListRolloutCampaignManagedDevices(ctx context.Context, rollout_campaignID string) ([]domain.ManagedDevice, error) {
	rows, err := p.pool.Query(ctx, `SELECT id,rollout_campaign_id,code,rollout_lane,required_successes,completed_installs FROM managed_devices WHERE rollout_campaign_id=$1 ORDER BY code`, rollout_campaignID)
	if err != nil {
		return nil, fmt.Errorf("list rollout_campaign managed_devices: %w", err)
	}
	defer rows.Close()
	items := make([]domain.ManagedDevice, 0)
	for rows.Next() {
		var item domain.ManagedDevice
		if err := rows.Scan(&item.ID, &item.RolloutCampaignID, &item.Code, &item.RolloutLane, &item.RequiredSuccesses, &item.Completed); err != nil {
			return nil, fmt.Errorf("scan managed_device: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
