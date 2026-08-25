package service

import (
	"context"
	"fmt"
	"time"

	"firmware-rollout-control/internal/rollout"
)

type FirmwareControl struct {
	store    *rollout.Store
	capacity *rollout.Capacity
	lease    rollout.Lease
}

func NewFirmwareControl(store *rollout.Store) *FirmwareControl {
	if store == nil {
		store = rollout.NewStore()
	}
	return &FirmwareControl{store: store, capacity: rollout.NewCapacity()}
}

func (c *FirmwareControl) Store() *rollout.Store { return c.store }

// SaveIdempotentResponse stores the response for an idempotency key when it is
// the first write and reports true (new). A repeated submit for an existing
// key replays the originally confirmed response instead of overwriting it, so
// it reports false (replay) — the second (possibly differing) content from a
// network retry never becomes the visible result, and later loads keep seeing
// the first confirmed wave.
func (c *FirmwareControl) SaveIdempotentResponse(tenantID, method, path, key string, response []byte) (bool, error) {
	scope, err := rollout.IdempotencyKey(tenantID, method, path, key)
	if err != nil {
		return false, err
	}
	return c.store.SaveIdempotent(scope, response), nil
}

func (c *FirmwareControl) LoadIdempotentResponse(tenantID, method, path, key string) ([]byte, bool, error) {
	scope, err := rollout.IdempotencyKey(tenantID, method, path, key)
	if err != nil {
		return nil, false, err
	}
	value, ok := c.store.Idempotent(scope)
	return value, ok, nil
}

func (c *FirmwareControl) RegisterCallback(callback rollout.Callback) (bool, error) {
	return c.store.RecordCallback(callback)
}

func (c *FirmwareControl) EnrollDevice(artifact rollout.Artifact, device rollout.Device) error {
	if err := rollout.ArtifactSupports(artifact, device); err != nil {
		return err
	}
	if err := c.store.PutArtifact(artifact); err != nil {
		return err
	}
	return c.store.PutDevice(device)
}

func (c *FirmwareControl) ApproveCampaign(campaign rollout.Campaign, artifact rollout.Artifact) error {
	if campaign.TenantID != artifact.TenantID || campaign.ArtifactID != artifact.ID || !artifact.Signed {
		return fmt.Errorf("campaign artifact approval is invalid: %w", rollout.ErrConflict)
	}
	campaign.ApprovedDigest = artifact.Digest
	return c.store.PutCampaign(campaign)
}

func (c *FirmwareControl) Dispatch(campaignID string, expectedVersion int64) error {
	return c.store.UpdateCampaign(campaignID, expectedVersion, func(campaign *rollout.Campaign) error {
		if err := rollout.CanDispatch(*campaign); err != nil {
			return err
		}
		campaign.State = "dispatching"
		return nil
	})
}

func (c *FirmwareControl) Promote(campaignID string, expectedVersion int64) error {
	return c.store.UpdateCampaign(campaignID, expectedVersion, func(campaign *rollout.Campaign) error {
		if err := rollout.CanPromote(*campaign); err != nil {
			return err
		}
		campaign.State = "promoted"
		return nil
	})
}

func (c *FirmwareControl) ReserveLane(tenantID, lane string, limit int) error {
	if tenantID == "" || lane == "" {
		return fmt.Errorf("rollout lane scope is missing: %w", rollout.ErrInvalid)
	}
	return c.capacity.Reserve(tenantID+"\x00"+lane, limit)
}

func (c *FirmwareControl) LaneUsage(tenantID, lane string) int {
	return c.capacity.Used(tenantID + "\x00" + lane)
}

func (c *FirmwareControl) AcquireWorker(owner string, now time.Time, ttl time.Duration) bool {
	return c.lease.Acquire(owner, now, ttl)
}

func (c *FirmwareControl) RenewWorker(owner string, now time.Time, ttl time.Duration) bool {
	return c.lease.Renew(owner, now, ttl)
}

func (c *FirmwareControl) ReleaseWorker(owner string) bool { return c.lease.Release(owner) }

func (c *FirmwareControl) Authenticate(token, tenantID, role string, now time.Time) error {
	session, ok := c.store.Session(token)
	if !ok {
		return fmt.Errorf("session not found: %w", rollout.ErrUnauthorized)
	}
	return rollout.AuthorizeSession(session, tenantID, role, now)
}

func (c *FirmwareControl) ChangeRole(token, nextRole string) error {
	session, ok := c.store.Session(token)
	if !ok {
		return fmt.Errorf("session not found: %w", rollout.ErrUnauthorized)
	}
	next, err := rollout.RotateRole(session, nextRole)
	if err != nil {
		return err
	}
	return c.store.PutSession(next)
}

func (c *FirmwareControl) Logout(token string, now time.Time) bool {
	return c.store.RevokeSession(token, now)
}

func (c *FirmwareControl) Devices(query rollout.Query) rollout.Page[rollout.Device] {
	return c.store.QueryDevices(query)
}

func (c *FirmwareControl) RecordAudit(event rollout.Event) error {
	return c.store.AppendEvent(event)
}

func (c *FirmwareControl) RunDownload(ctx context.Context, timeout time.Duration, download func(context.Context) error) error {
	operationCtx, cancel := rollout.DerivedOperationContext(ctx, timeout)
	defer cancel()
	return download(operationCtx)
}

func (c *FirmwareControl) WaitRetry(ctx context.Context, delay time.Duration) error {
	return rollout.WaitBackoff(ctx, delay)
}

func (c *FirmwareControl) SnapshotArtifact(id string) (rollout.Artifact, bool) {
	return c.store.Artifact(id)
}

func (c *FirmwareControl) SnapshotCampaign(id string) (rollout.Campaign, bool) {
	return c.store.Campaign(id)
}

func (c *FirmwareControl) RestoreArtifactLabels(snapshot map[string]string) map[string]string {
	return rollout.RestoreLabels(snapshot)
}

func (c *FirmwareControl) ErrorResponse(err error) (int, string) {
	return rollout.ClassifyError(err)
}

func (c *FirmwareControl) CloseSafetyAlert(campaignID string, openAlerts int) error {
	campaign, ok := c.store.Campaign(campaignID)
	if !ok {
		return fmt.Errorf("campaign not found: %w", rollout.ErrConflict)
	}
	return rollout.CanCloseAlert(campaign, openAlerts)
}

func (c *FirmwareControl) RollbackTarget(deviceID string, expectedGeneration int64) (string, error) {
	device, ok := c.store.Device(deviceID)
	if !ok {
		return "", fmt.Errorf("device not found: %w", rollout.ErrConflict)
	}
	if err := rollout.CheckGeneration(device, expectedGeneration); err != nil {
		return "", err
	}
	return rollout.RollbackVersion(device)
}
