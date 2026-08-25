package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"firmware-rollout-control/internal/rollout"
	"firmware-rollout-control/internal/service"
)

func newFirmwareControl() (*service.FirmwareControl, *rollout.Store) {
	store := rollout.NewStore()
	return service.NewFirmwareControl(store), store
}

func signedArtifact() rollout.Artifact {
	return rollout.Artifact{ID: "artifact-1", TenantID: "tenant-a", Version: "2.4.0", Digest: "sha256:verified", DeviceClasses: []string{"edge-v2"}, Labels: map[string]string{"ring": "canary"}, Signed: true}
}

func TestIdempotencySeparatesTenants(t *testing.T) {
	control, _ := newFirmwareControl()
	if stored, err := control.SaveIdempotentResponse("tenant-a", "POST", "/campaigns", "request-7", []byte(`{"id":"a"}`)); err != nil || !stored {
		t.Fatalf("first response stored=%v err=%v", stored, err)
	}
	if stored, err := control.SaveIdempotentResponse("tenant-b", "POST", "/campaigns", "request-7", []byte(`{"id":"b"}`)); err != nil || !stored {
		t.Fatalf("second tenant stored=%v err=%v", stored, err)
	}
	value, ok, err := control.LoadIdempotentResponse("tenant-b", "POST", "/campaigns", "request-7")
	if err != nil || !ok || string(value) != `{"id":"b"}` {
		t.Fatalf("tenant-b response=%s ok=%v err=%v", value, ok, err)
	}
}

func TestIdempotencySeparatesRoutes(t *testing.T) {
	control, _ := newFirmwareControl()
	for _, path := range []string{"/campaigns", "/artifacts"} {
		stored, err := control.SaveIdempotentResponse("tenant-a", "POST", path, "request-8", []byte(path))
		if err != nil || !stored {
			t.Fatalf("path=%s stored=%v err=%v", path, stored, err)
		}
	}
	value, ok, err := control.LoadIdempotentResponse("tenant-a", "POST", "/artifacts", "request-8")
	if err != nil || !ok || string(value) != "/artifacts" {
		t.Fatalf("route response=%s ok=%v err=%v", value, ok, err)
	}
}

func TestIdempotencyReplaysFirstResponseOnRetriedSubmit(t *testing.T) {
	control, _ := newFirmwareControl()
	first := []byte(`{"id":"wave-1"}`)
	if stored, err := control.SaveIdempotentResponse("tenant-a", "POST", "/v1/rollout_waves", "idem-retry", first); err != nil || !stored {
		t.Fatalf("first submit stored=%v err=%v", stored, err)
	}
	// A network retry resubmits with the same key but different response content.
	if stored, err := control.SaveIdempotentResponse("tenant-a", "POST", "/v1/rollout_waves", "idem-retry", []byte(`{"id":"wave-2"}`)); err != nil || stored {
		t.Fatalf("duplicate submit stored=%v err=%v, want false (replay)", stored, err)
	}
	value, ok, err := control.LoadIdempotentResponse("tenant-a", "POST", "/v1/rollout_waves", "idem-retry")
	if err != nil || !ok || string(value) != string(first) {
		t.Fatalf("replayed response=%s ok=%v err=%v, want first %s", value, ok, err, first)
	}
}

func TestCallbackIdentityIncludesArtifact(t *testing.T) {
	control, _ := newFirmwareControl()
	base := rollout.Callback{TenantID: "tenant-a", DeviceID: "device-1", ArtifactID: "artifact-1", EventID: "callback-9", Status: "installed"}
	first, err := control.RegisterCallback(base)
	if err != nil || !first {
		t.Fatalf("first callback new=%v err=%v", first, err)
	}
	base.ArtifactID = "artifact-2"
	second, err := control.RegisterCallback(base)
	if err != nil || !second {
		t.Fatalf("second artifact callback new=%v err=%v", second, err)
	}
}

func TestEnrollmentRejectsCrossTenantArtifact(t *testing.T) {
	control, store := newFirmwareControl()
	artifact := signedArtifact()
	device := rollout.Device{ID: "device-1", TenantID: "tenant-b", Class: "edge-v2"}
	if err := control.EnrollDevice(artifact, device); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("enrollment err=%v", err)
	}
	if _, ok := store.Device(device.ID); ok {
		t.Fatal("cross-tenant device was persisted")
	}
}

func TestEnrollmentRejectsUnsupportedClass(t *testing.T) {
	control, store := newFirmwareControl()
	artifact := signedArtifact()
	device := rollout.Device{ID: "device-2", TenantID: artifact.TenantID, Class: "legacy-v1"}
	if err := control.EnrollDevice(artifact, device); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("enrollment err=%v", err)
	}
	if _, ok := store.Artifact(artifact.ID); ok {
		t.Fatal("artifact was persisted before compatibility passed")
	}
}

func TestEnrollmentRequiresVerifiedSignature(t *testing.T) {
	control, store := newFirmwareControl()
	artifact := signedArtifact()
	artifact.Signed = false
	device := rollout.Device{ID: "device-3", TenantID: artifact.TenantID, Class: "edge-v2"}
	if err := control.EnrollDevice(artifact, device); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("enrollment err=%v", err)
	}
	if _, ok := store.Device(device.ID); ok {
		t.Fatal("device was enrolled with an unverified artifact")
	}
}

func TestApprovalPinsArtifactDigest(t *testing.T) {
	control, _ := newFirmwareControl()
	artifact := signedArtifact()
	campaign := rollout.Campaign{ID: "campaign-1", TenantID: artifact.TenantID, ArtifactID: artifact.ID, State: "running", Version: 1}
	if err := control.ApproveCampaign(campaign, artifact); err != nil {
		t.Fatal(err)
	}
	snapshot, ok := control.SnapshotCampaign(campaign.ID)
	if !ok || snapshot.ApprovedDigest != artifact.Digest {
		t.Fatalf("approved digest=%q ok=%v", snapshot.ApprovedDigest, ok)
	}
}

func TestPausedCampaignCannotDispatch(t *testing.T) {
	control, store := newFirmwareControl()
	campaign := rollout.Campaign{ID: "campaign-2", TenantID: "tenant-a", ArtifactID: "artifact-1", ApprovedDigest: "sha256:ok", State: "paused", Version: 4}
	if err := store.PutCampaign(campaign); err != nil {
		t.Fatal(err)
	}
	if err := control.Dispatch(campaign.ID, campaign.Version); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("dispatch err=%v", err)
	}
	snapshot, _ := store.Campaign(campaign.ID)
	if snapshot.State != "paused" || snapshot.Version != campaign.Version {
		t.Fatalf("campaign changed=%+v", snapshot)
	}
}

func TestPromotionRequiresHealthyCanaries(t *testing.T) {
	control, store := newFirmwareControl()
	campaign := rollout.Campaign{ID: "campaign-3", TenantID: "tenant-a", State: "running", RequiredHealthy: 3, Healthy: 2, Version: 1}
	if err := store.PutCampaign(campaign); err != nil {
		t.Fatal(err)
	}
	if err := control.Promote(campaign.ID, campaign.Version); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("promotion err=%v", err)
	}
	snapshot, _ := store.Campaign(campaign.ID)
	if snapshot.State != "running" {
		t.Fatalf("campaign state=%s", snapshot.State)
	}
}

func TestPromotionRejectsFailedCanary(t *testing.T) {
	control, store := newFirmwareControl()
	campaign := rollout.Campaign{ID: "campaign-4", TenantID: "tenant-a", State: "running", RequiredHealthy: 2, Healthy: 2, Failed: 1, Version: 6}
	if err := store.PutCampaign(campaign); err != nil {
		t.Fatal(err)
	}
	if err := control.Promote(campaign.ID, campaign.Version); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("promotion err=%v", err)
	}
}

func TestRollbackUsesPreviousVersion(t *testing.T) {
	control, store := newFirmwareControl()
	device := rollout.Device{ID: "device-4", TenantID: "tenant-a", CurrentVersion: "2.4.0", PreviousVersion: "2.3.7", Generation: 9}
	if err := store.PutDevice(device); err != nil {
		t.Fatal(err)
	}
	target, err := control.RollbackTarget(device.ID, device.Generation)
	if err != nil || target != device.PreviousVersion {
		t.Fatalf("rollback target=%q err=%v", target, err)
	}
}

func TestRollbackRejectsStaleGeneration(t *testing.T) {
	control, store := newFirmwareControl()
	device := rollout.Device{ID: "device-5", TenantID: "tenant-a", CurrentVersion: "2.4.0", PreviousVersion: "2.3.7", Generation: 12}
	if err := store.PutDevice(device); err != nil {
		t.Fatal(err)
	}
	if _, err := control.RollbackTarget(device.ID, 11); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("rollback err=%v", err)
	}
}

func TestSafetyAlertStaysOpenWhileAlertsRemain(t *testing.T) {
	control, store := newFirmwareControl()
	campaign := rollout.Campaign{ID: "campaign-5", TenantID: "tenant-a", State: "paused"}
	if err := store.PutCampaign(campaign); err != nil {
		t.Fatal(err)
	}
	if err := control.CloseSafetyAlert(campaign.ID, 2); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("close err=%v", err)
	}
}

func TestSafetyAlertStaysOpenDuringRollback(t *testing.T) {
	control, store := newFirmwareControl()
	campaign := rollout.Campaign{ID: "campaign-6", TenantID: "tenant-a", State: "rolling_back"}
	if err := store.PutCampaign(campaign); err != nil {
		t.Fatal(err)
	}
	if err := control.CloseSafetyAlert(campaign.ID, 0); !errors.Is(err, rollout.ErrConflict) {
		t.Fatalf("close err=%v", err)
	}
}

func TestArtifactSnapshotDoesNotShareLabels(t *testing.T) {
	control, store := newFirmwareControl()
	artifact := signedArtifact()
	if err := store.PutArtifact(artifact); err != nil {
		t.Fatal(err)
	}
	snapshot, _ := control.SnapshotArtifact(artifact.ID)
	snapshot.Labels["ring"] = "fleet"
	again, _ := control.SnapshotArtifact(artifact.ID)
	if again.Labels["ring"] != "canary" || artifact.Labels["ring"] != "canary" {
		t.Fatalf("labels leaked: stored=%v input=%v", again.Labels, artifact.Labels)
	}
}

func TestArtifactSnapshotDoesNotShareClasses(t *testing.T) {
	control, store := newFirmwareControl()
	artifact := signedArtifact()
	if err := store.PutArtifact(artifact); err != nil {
		t.Fatal(err)
	}
	snapshot, _ := control.SnapshotArtifact(artifact.ID)
	snapshot.DeviceClasses[0] = "legacy-v1"
	again, _ := control.SnapshotArtifact(artifact.ID)
	if again.DeviceClasses[0] != "edge-v2" || artifact.DeviceClasses[0] != "edge-v2" {
		t.Fatalf("classes leaked: stored=%v input=%v", again.DeviceClasses, artifact.DeviceClasses)
	}
}

func TestCampaignSnapshotDoesNotShareDevices(t *testing.T) {
	control, store := newFirmwareControl()
	campaign := rollout.Campaign{ID: "campaign-7", TenantID: "tenant-a", DeviceIDs: []string{"device-1", "device-2"}}
	if err := store.PutCampaign(campaign); err != nil {
		t.Fatal(err)
	}
	snapshot, _ := control.SnapshotCampaign(campaign.ID)
	snapshot.DeviceIDs[0] = "device-x"
	again, _ := control.SnapshotCampaign(campaign.ID)
	if again.DeviceIDs[0] != "device-1" || campaign.DeviceIDs[0] != "device-1" {
		t.Fatalf("devices leaked: stored=%v input=%v", again.DeviceIDs, campaign.DeviceIDs)
	}
}

func TestRestoredLabelsAreWritableAndIsolated(t *testing.T) {
	control, _ := newFirmwareControl()
	snapshot := map[string]string{"ring": "canary"}
	restored := control.RestoreArtifactLabels(snapshot)
	restored["region"] = "north"
	restored["ring"] = "fleet"
	if snapshot["ring"] != "canary" || snapshot["region"] != "" {
		t.Fatalf("snapshot was mutated: %v", snapshot)
	}
}

func TestRetryWaitStopsOnCancellation(t *testing.T) {
	control, _ := newFirmwareControl()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	err := control.WaitRetry(ctx, time.Second)
	if !errors.Is(err, context.Canceled) || time.Since(started) > 200*time.Millisecond {
		t.Fatalf("wait err=%v elapsed=%s", err, time.Since(started))
	}
}

func TestDownloadContextKeepsParentCancellation(t *testing.T) {
	control, _ := newFirmwareControl()
	parent, cancel := context.WithCancel(context.Background())
	cancel()
	err := control.RunDownload(parent, time.Minute, func(ctx context.Context) error { return ctx.Err() })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("download err=%v", err)
	}
}

func TestWorkerLeaseRejectsForeignRelease(t *testing.T) {
	control, _ := newFirmwareControl()
	now := time.Now()
	if !control.AcquireWorker("worker-a", now, time.Minute) {
		t.Fatal("worker-a did not acquire lease")
	}
	if control.ReleaseWorker("worker-b") {
		t.Fatal("worker-b released worker-a lease")
	}
	if control.AcquireWorker("worker-b", now.Add(time.Second), time.Minute) {
		t.Fatal("worker-b acquired an active foreign lease")
	}
}

func TestExpiredWorkerLeaseCannotRenew(t *testing.T) {
	control, _ := newFirmwareControl()
	now := time.Now()
	if !control.AcquireWorker("worker-a", now, time.Second) {
		t.Fatal("worker-a did not acquire lease")
	}
	if control.RenewWorker("worker-a", now.Add(time.Second), time.Minute) {
		t.Fatal("expired lease was renewed")
	}
	if !control.AcquireWorker("worker-b", now.Add(time.Second), time.Minute) {
		t.Fatal("worker-b could not take expired lease")
	}
}

func TestConcurrentLaneCapacityDoesNotOversubscribe(t *testing.T) {
	control, _ := newFirmwareControl()
	start := make(chan struct{})
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			results <- control.ReserveLane("tenant-a", "north", 1)
		}()
	}
	ready.Wait()
	close(start)
	var successes int
	for range 2 {
		if err := <-results; err == nil {
			successes++
		}
	}
	if successes != 1 || control.LaneUsage("tenant-a", "north") != 1 {
		t.Fatalf("successes=%d used=%d", successes, control.LaneUsage("tenant-a", "north"))
	}
}

func TestSessionCannotCrossTenant(t *testing.T) {
	control, store := newFirmwareControl()
	now := time.Now()
	if err := store.PutSession(rollout.Session{Token: "token-1", TenantID: "tenant-a", UserID: "operator-1", Role: "release_manager", ExpiresAt: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := control.Authenticate("token-1", "tenant-b", "release_manager", now); !errors.Is(err, rollout.ErrUnauthorized) {
		t.Fatalf("authentication err=%v", err)
	}
}

func TestExpiredSessionCannotAuthorize(t *testing.T) {
	control, store := newFirmwareControl()
	now := time.Now()
	if err := store.PutSession(rollout.Session{Token: "token-2", TenantID: "tenant-a", UserID: "operator-2", Role: "release_manager", ExpiresAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := control.Authenticate("token-2", "tenant-a", "release_manager", now); !errors.Is(err, rollout.ErrUnauthorized) {
		t.Fatalf("authentication err=%v", err)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	control, store := newFirmwareControl()
	now := time.Now()
	if err := store.PutSession(rollout.Session{Token: "token-3", TenantID: "tenant-a", UserID: "operator-3", Role: "release_manager", ExpiresAt: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if !control.Logout("token-3", now) {
		t.Fatal("logout did not revoke token")
	}
	if err := control.Authenticate("token-3", "tenant-a", "release_manager", now); !errors.Is(err, rollout.ErrUnauthorized) {
		t.Fatalf("authentication err=%v", err)
	}
}

func TestRoleChangeUpdatesAuthorization(t *testing.T) {
	control, store := newFirmwareControl()
	now := time.Now()
	if err := store.PutSession(rollout.Session{Token: "token-4", TenantID: "tenant-a", UserID: "operator-4", Role: "release_manager", ExpiresAt: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := control.ChangeRole("token-4", "auditor"); err != nil {
		t.Fatal(err)
	}
	if err := control.Authenticate("token-4", "tenant-a", "release_manager", now); !errors.Is(err, rollout.ErrUnauthorized) {
		t.Fatalf("old role authentication err=%v", err)
	}
	if err := control.Authenticate("token-4", "tenant-a", "auditor", now); err != nil {
		t.Fatalf("new role authentication err=%v", err)
	}
}

func TestDeviceQueryDoesNotCrossTenant(t *testing.T) {
	control, store := newFirmwareControl()
	for _, device := range []rollout.Device{{ID: "device-a", TenantID: "tenant-a", Class: "edge-v2"}, {ID: "device-b", TenantID: "tenant-b", Class: "edge-v2"}} {
		if err := store.PutDevice(device); err != nil {
			t.Fatal(err)
		}
	}
	page := control.Devices(rollout.Query{TenantID: "tenant-a", Limit: 10})
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].TenantID != "tenant-a" {
		t.Fatalf("page=%+v", page)
	}
}

func TestDeviceQueryTotalUsesClassFilter(t *testing.T) {
	control, store := newFirmwareControl()
	for _, device := range []rollout.Device{{ID: "device-a", TenantID: "tenant-a", Class: "edge-v2"}, {ID: "device-b", TenantID: "tenant-a", Class: "gateway-v3"}} {
		if err := store.PutDevice(device); err != nil {
			t.Fatal(err)
		}
	}
	page := control.Devices(rollout.Query{TenantID: "tenant-a", Class: "edge-v2", Limit: 10})
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Class != "edge-v2" {
		t.Fatalf("page=%+v", page)
	}
}

func TestSignatureErrorKeepsHTTPClassification(t *testing.T) {
	control, _ := newFirmwareControl()
	err := rollout.WrapOperation("publish", &rollout.SignatureError{Digest: "sha256:bad", Cause: rollout.ErrConflict})
	status, code := control.ErrorResponse(err)
	if status != 422 || code != "signature_invalid" {
		t.Fatalf("status=%d code=%s err=%v", status, code, err)
	}
}
