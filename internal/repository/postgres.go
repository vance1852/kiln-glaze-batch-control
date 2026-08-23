package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"firmware-rollout-control/internal/db"
	"firmware-rollout-control/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Postgres struct {
	pool *db.Pool
}

func NewPostgres(pool *db.Pool) *Postgres { return &Postgres{pool: pool} }

func (p *Postgres) InTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(&transaction{tx: tx}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (p *Postgres) Close() error { p.pool.Close(); return nil }

func (p *Postgres) CreateRolloutCampaign(ctx context.Context, rollout_campaign *domain.RolloutCampaign) error {
	return createRolloutCampaign(ctx, p.pool, rollout_campaign)
}

func (p *Postgres) GetRolloutCampaign(ctx context.Context, id string) (domain.RolloutCampaign, error) {
	return getRolloutCampaign(ctx, p.pool, id)
}

func (p *Postgres) AdvanceRolloutCampaign(ctx context.Context, id string, status domain.RolloutCampaignStatus, version int64) error {
	return advanceRolloutCampaign(ctx, p.pool, id, status, version)
}

func (p *Postgres) CreateManagedDevice(ctx context.Context, in ManagedDeviceInput) (string, error) {
	return createManagedDevice(ctx, p.pool, in)
}
func (p *Postgres) CreateDeploymentJob(ctx context.Context, in DeploymentJobInput) (domain.DeploymentJob, error) {
	return createDeploymentJob(ctx, p.pool, in)
}

func (p *Postgres) CreateDeploymentJobDetached(ctx context.Context, in DeploymentJobInput) (domain.DeploymentJob, error) {
	return createDeploymentJob(ctx, p.pool, in)
}
func (p *Postgres) GetDeploymentJob(ctx context.Context, id string) (domain.DeploymentJob, error) {
	return getDeploymentJob(ctx, p.pool, id)
}
func (p *Postgres) MoveDeploymentJob(ctx context.Context, id string, status domain.DeploymentJobStatus, version int64, now time.Time) error {
	return moveDeploymentJob(ctx, p.pool, id, status, version, now)
}
func (p *Postgres) RecordActivation(ctx context.Context, in ActivationInput) error {
	return recordActivation(ctx, p.pool, in)
}
func (p *Postgres) CreateRolloutWave(ctx context.Context, in RolloutWaveInput) (string, error) {
	return createRolloutWave(ctx, p.pool, in)
}
func (p *Postgres) AttachDeploymentJobs(ctx context.Context, rollout_waveID string, taskIDs []string) error {
	return attachDeploymentJobs(ctx, p.pool, rollout_waveID, taskIDs)
}
func (p *Postgres) CreateInstallationReport(ctx context.Context, in InstallationReportInput) (string, error) {
	return createInstallationReport(ctx, p.pool, in)
}
func (p *Postgres) ReviewInstallationReportRecord(ctx context.Context, id string, accepted bool, version int64, now time.Time) error {
	return reviewInstallationReportRecord(ctx, p.pool, id, accepted, version, now)
}
func (p *Postgres) CreateHealthAlert(ctx context.Context, in HealthAlertInput) (string, error) {
	return createHealthAlert(ctx, p.pool, in)
}
func (p *Postgres) ListDeploymentJobs(ctx context.Context, offset, limit int, rollout_campaignID string, status domain.DeploymentJobStatus) (Page, error) {
	return listDeploymentJobs(ctx, p.pool, offset, limit, rollout_campaignID, status)
}
func (p *Postgres) DueHealthAlerts(ctx context.Context, before time.Time, limit int) ([]HealthAlertInput, error) {
	return dueHealthAlerts(ctx, p.pool, before, limit)
}
func (p *Postgres) WriteAudit(ctx context.Context, in AuditInput) error {
	return writeAudit(ctx, p.pool, in)
}

type transaction struct{ tx pgx.Tx }

func (t *transaction) InTx(_ context.Context, _ func(Repository) error) error {
	return errors.New("nested transaction")
}
func (t *transaction) Close() error { return nil }
func (t *transaction) CreateRolloutCampaign(ctx context.Context, p *domain.RolloutCampaign) error {
	return createRolloutCampaign(ctx, t.tx, p)
}
func (t *transaction) GetRolloutCampaign(ctx context.Context, id string) (domain.RolloutCampaign, error) {
	return getRolloutCampaign(ctx, t.tx, id)
}
func (t *transaction) AdvanceRolloutCampaign(ctx context.Context, id string, s domain.RolloutCampaignStatus, v int64) error {
	return advanceRolloutCampaign(ctx, t.tx, id, s, v)
}
func (t *transaction) CreateManagedDevice(ctx context.Context, in ManagedDeviceInput) (string, error) {
	return createManagedDevice(ctx, t.tx, in)
}
func (t *transaction) CreateDeploymentJob(ctx context.Context, in DeploymentJobInput) (domain.DeploymentJob, error) {
	return createDeploymentJob(ctx, t.tx, in)
}
func (t *transaction) GetDeploymentJob(ctx context.Context, id string) (domain.DeploymentJob, error) {
	return getDeploymentJob(ctx, t.tx, id)
}
func (t *transaction) MoveDeploymentJob(ctx context.Context, id string, s domain.DeploymentJobStatus, v int64, now time.Time) error {
	return moveDeploymentJob(ctx, t.tx, id, s, v, now)
}
func (t *transaction) RecordActivation(ctx context.Context, in ActivationInput) error {
	return recordActivation(ctx, t.tx, in)
}
func (t *transaction) CreateRolloutWave(ctx context.Context, in RolloutWaveInput) (string, error) {
	return createRolloutWave(ctx, t.tx, in)
}
func (t *transaction) AttachDeploymentJobs(ctx context.Context, id string, ids []string) error {
	return attachDeploymentJobs(ctx, t.tx, id, ids)
}
func (t *transaction) CreateInstallationReport(ctx context.Context, in InstallationReportInput) (string, error) {
	return createInstallationReport(ctx, t.tx, in)
}
func (t *transaction) ReviewInstallationReportRecord(ctx context.Context, id string, accepted bool, v int64, now time.Time) error {
	return reviewInstallationReportRecord(ctx, t.tx, id, accepted, v, now)
}
func (t *transaction) CreateHealthAlert(ctx context.Context, in HealthAlertInput) (string, error) {
	return createHealthAlert(ctx, t.tx, in)
}
func (t *transaction) ListDeploymentJobs(ctx context.Context, offset, limit int, rollout_campaignID string, status domain.DeploymentJobStatus) (Page, error) {
	return listDeploymentJobs(ctx, t.tx, offset, limit, rollout_campaignID, status)
}
func (t *transaction) DueHealthAlerts(ctx context.Context, before time.Time, limit int) ([]HealthAlertInput, error) {
	return dueHealthAlerts(ctx, t.tx, before, limit)
}
func (t *transaction) WriteAudit(ctx context.Context, in AuditInput) error {
	return writeAudit(ctx, t.tx, in)
}

type sqler interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func createRolloutCampaign(ctx context.Context, q sqler, rollout_campaign *domain.RolloutCampaign) error {
	_, err := q.Exec(ctx, `INSERT INTO rollout_campaigns(id,code,name,status,timezone,starts_at,ends_at,version,created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, rollout_campaign.ID, rollout_campaign.Code, rollout_campaign.Name, rollout_campaign.Status, rollout_campaign.Timezone, rollout_campaign.StartsAt, rollout_campaign.EndsAt, rollout_campaign.Version, rollout_campaign.CreatedBy)
	return wrapWrite(err)
}

func getRolloutCampaign(ctx context.Context, q sqler, id string) (domain.RolloutCampaign, error) {
	var p domain.RolloutCampaign
	err := q.QueryRow(ctx, `SELECT id,code,name,status,timezone,starts_at,ends_at,version,created_by FROM rollout_campaigns WHERE id=$1`, id).Scan(&p.ID, &p.Code, &p.Name, &p.Status, &p.Timezone, &p.StartsAt, &p.EndsAt, &p.Version, &p.CreatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RolloutCampaign{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.RolloutCampaign{}, fmt.Errorf("get rollout_campaign: %w", err)
	}
	return p, nil
}

func advanceRolloutCampaign(ctx context.Context, q sqler, id string, status domain.RolloutCampaignStatus, version int64) error {
	result, err := q.Exec(ctx, `UPDATE rollout_campaigns SET status=$1,version=version+1 WHERE id=$2 AND version=$3`, status, id, version)
	if err != nil {
		return fmt.Errorf("advance rollout_campaign: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func createManagedDevice(ctx context.Context, q sqler, in ManagedDeviceInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO managed_devices(id,rollout_campaign_id,code,rollout_lane,required_successes) VALUES ($1,$2,$3,$4,$5)`, id, in.RolloutCampaignID, in.Code, in.RolloutLane, in.RequiredSuccesses)
	return wrapID(err, id)
}

func createDeploymentJob(ctx context.Context, q sqler, in DeploymentJobInput) (domain.DeploymentJob, error) {
	s := domain.DeploymentJob{ID: uuid.NewString(), RolloutCampaignID: in.RolloutCampaignID, ManagedDeviceID: in.ManagedDeviceID, TaskCode: in.TaskCode, Status: domain.DeploymentJobQueued, ExpiresAt: in.ExpiresAt, Version: 1}
	_, err := q.Exec(ctx, `INSERT INTO deployment_jobs(id,rollout_campaign_id,managed_device_id,task_code,status,expires_at,version) VALUES ($1,$2,$3,$4,$5,$6,$7)`, s.ID, s.RolloutCampaignID, s.ManagedDeviceID, s.TaskCode, s.Status, s.ExpiresAt, s.Version)
	return wrapDeploymentJob(err, s)
}

func getDeploymentJob(ctx context.Context, q sqler, id string) (domain.DeploymentJob, error) {
	var s domain.DeploymentJob
	err := q.QueryRow(ctx, `SELECT id,rollout_campaign_id,managed_device_id,task_code,status,completed_at,accepted_at,expires_at,version FROM deployment_jobs WHERE id=$1`, id).Scan(&s.ID, &s.RolloutCampaignID, &s.ManagedDeviceID, &s.TaskCode, &s.Status, &s.CompletedAt, &s.AcceptedAt, &s.ExpiresAt, &s.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DeploymentJob{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DeploymentJob{}, fmt.Errorf("get task: %w", err)
	}
	return s, nil
}

func moveDeploymentJob(ctx context.Context, q sqler, id string, status domain.DeploymentJobStatus, version int64, now time.Time) error {
	if status == domain.DeploymentJobCompleted {
		result, err := q.Exec(ctx, `WITH eligible_managed_device AS (
			SELECT id FROM managed_devices WHERE id=(SELECT managed_device_id FROM deployment_jobs WHERE id=$3) AND completed_installs < required_successes FOR UPDATE
		), updated AS (
			UPDATE deployment_jobs SET status=$1,version=version+1,completed_at=$2 WHERE id=$3 AND version=$4 AND EXISTS (SELECT 1 FROM eligible_managed_device) RETURNING managed_device_id
		)
		UPDATE managed_devices SET completed_installs=completed_installs+1 WHERE id IN (SELECT managed_device_id FROM updated)`, status, now, id, version)
		if err != nil {
			return fmt.Errorf("complete task: %w", err)
		}
		if result.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		return nil
	}
	result, err := q.Exec(ctx, `UPDATE deployment_jobs SET status=$1,version=version+1,completed_at=CASE WHEN $1='completed' THEN $2 ELSE completed_at END,accepted_at=CASE WHEN $1='accepted' THEN $2 ELSE accepted_at END WHERE id=$3 AND version=$4`, status, now, id, version)
	if err != nil {
		return fmt.Errorf("move task: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func recordActivation(ctx context.Context, q sqler, in ActivationInput) error {
	_, err := q.Exec(ctx, `INSERT INTO activation_events(id,deployment_job_id,from_operator,to_operator,location,recorded_at,note) VALUES ($1,$2,$3,$4,$5,$6,$7)`, uuid.NewString(), in.DeploymentJobID, in.From, in.To, in.Location, in.RecordedAt, in.Note)
	return wrapWrite(err)
}

func createRolloutWave(ctx context.Context, q sqler, in RolloutWaveInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO rollout_waves(id,code,status,method,capacity) VALUES ($1,$2,'queued',$3,$4)`, id, in.Code, in.Method, in.Capacity)
	return wrapID(err, id)
}

func attachDeploymentJobs(ctx context.Context, q sqler, rollout_waveID string, taskIDs []string) error {
	for _, taskID := range taskIDs {
		if _, err := q.Exec(ctx, `INSERT INTO rollout_wave_items(rollout_wave_id,deployment_job_id) VALUES ($1,$2)`, rollout_waveID, taskID); err != nil {
			return wrapWrite(err)
		}
		result, err := q.Exec(ctx, `UPDATE deployment_jobs SET status='in_progress',version=version+1 WHERE id=$1 AND status='accepted' AND expires_at >= now()`, taskID)
		if err != nil {
			return wrapWrite(err)
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("task %s is not eligible for managed_device-round execution: %w", taskID, domain.ErrInvalidTransition)
		}
	}
	return nil
}

func createInstallationReport(ctx context.Context, q sqler, in InstallationReportInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO installation_reports(id,deployment_job_id,rollout_wave_id,recorded_by,status,risk_score,scale,alert_threshold,observed_at) VALUES ($1,$2,$3,$4,'pending',$5,$6,$7,$8)`, id, in.DeploymentJobID, in.RolloutWaveID, in.RecorderID, in.RiskScore, in.Scale, in.AlertThreshold, in.ObservedAt)
	return wrapID(err, id)
}

func reviewInstallationReportRecord(ctx context.Context, q sqler, id string, accepted bool, version int64, now time.Time) error {
	status := domain.InstallationReportVerified
	if !accepted {
		status = domain.InstallationReportRejected
	}
	result, err := q.Exec(ctx, `UPDATE installation_reports SET status=$1,reviewed_at=$2,version=version+1 WHERE id=$3 AND status='pending' AND version=$4`, status, now, id, version)
	if err != nil {
		return fmt.Errorf("review installation_report: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func createHealthAlert(ctx context.Context, q sqler, in HealthAlertInput) (string, error) {
	id := uuid.NewString()
	_, err := q.Exec(ctx, `INSERT INTO health_alerts(id,deployment_job_id,kind,status,reason,due_at) VALUES ($1,$2,$3,'open',$4,$5)`, id, in.DeploymentJobID, in.Kind, in.Reason, in.DueAt)
	return wrapID(err, id)
}

func listDeploymentJobs(ctx context.Context, q sqler, offset, limit int, rollout_campaignID string, status domain.DeploymentJobStatus) (Page, error) {
	page := Page{Offset: offset, Limit: limit, Items: make([]domain.DeploymentJob, 0)}
	args := []any{limit, offset}
	where := "WHERE TRUE"
	if rollout_campaignID != "" {
		args = append(args, rollout_campaignID)
		where += fmt.Sprintf(" AND rollout_campaign_id=$%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	query := fmt.Sprintf(`SELECT id,rollout_campaign_id,managed_device_id,task_code,status,completed_at,accepted_at,expires_at,version FROM deployment_jobs %s ORDER BY created_at DESC LIMIT $1 OFFSET $2`, where)
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return Page{}, fmt.Errorf("list deployment_jobs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s domain.DeploymentJob
		if err := rows.Scan(&s.ID, &s.RolloutCampaignID, &s.ManagedDeviceID, &s.TaskCode, &s.Status, &s.CompletedAt, &s.AcceptedAt, &s.ExpiresAt, &s.Version); err != nil {
			return Page{}, fmt.Errorf("scan task: %w", err)
		}
		page.Items = append(page.Items, s)
	}
	if err := rows.Err(); err != nil {
		return Page{}, fmt.Errorf("list task rows: %w", err)
	}
	countWhere := "WHERE TRUE"
	countArgs := make([]any, 0, len(args)-2)
	if rollout_campaignID != "" {
		countArgs = append(countArgs, rollout_campaignID)
		countWhere += fmt.Sprintf(" AND rollout_campaign_id=$%d", len(countArgs))
	}
	if status != "" {
		countArgs = append(countArgs, status)
		countWhere += fmt.Sprintf(" AND status=$%d", len(countArgs))
	}
	countQuery := fmt.Sprintf("SELECT count(*) FROM deployment_jobs %s", countWhere)
	if err := q.QueryRow(ctx, countQuery, countArgs...).Scan(&page.Total); err != nil {
		return Page{}, fmt.Errorf("count deployment_jobs: %w", err)
	}
	return page, nil
}

func dueHealthAlerts(ctx context.Context, q sqler, before time.Time, limit int) ([]HealthAlertInput, error) {
	rows, err := q.Query(ctx, `SELECT deployment_job_id,kind,reason,due_at FROM health_alerts WHERE status IN ('open','in_progress') AND due_at <= $1 ORDER BY due_at LIMIT $2`, before, limit)
	if err != nil {
		return nil, fmt.Errorf("due health_alerts: %w", err)
	}
	defer rows.Close()
	out := make([]HealthAlertInput, 0)
	for rows.Next() {
		var item HealthAlertInput
		if err := rows.Scan(&item.DeploymentJobID, &item.Kind, &item.Reason, &item.DueAt); err != nil {
			return nil, fmt.Errorf("scan safety_alert: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func writeAudit(ctx context.Context, q sqler, in AuditInput) error {
	detail := in.Detail
	if len(detail) == 0 {
		detail = []byte(`{}`)
	}
	if !json.Valid(detail) {
		return fmt.Errorf("invalid audit detail: %w", domain.ErrConflict)
	}
	_, err := q.Exec(ctx, `INSERT INTO audit_events(id,request_id,release_operator_id,object_type,object_id,action,outcome,detail) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, uuid.NewString(), in.RequestID, in.ReleaseOperatorID, in.ObjectType, in.ObjectID, in.Action, in.Outcome, detail)
	return wrapWrite(err)
}

func wrapWrite(err error) error {
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "22003", "22P02", "23502", "23503", "23505", "23514":
			return fmt.Errorf("repository write: %w: %w", err, domain.ErrConflict)
		}
	}
	return fmt.Errorf("repository write: %w", err)
}
func wrapID(err error, id string) (string, error) {
	if err != nil {
		return "", wrapWrite(err)
	}
	return id, nil
}
func wrapDeploymentJob(err error, s domain.DeploymentJob) (domain.DeploymentJob, error) {
	if err != nil {
		return domain.DeploymentJob{}, wrapWrite(err)
	}
	return s, nil
}
