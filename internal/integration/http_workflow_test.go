package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/httpapi"
	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
)

func apiRequest(t *testing.T, handler http.Handler, method, path string, body any, release_operatorID string, wantStatus int, target any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "http-contract")
	if release_operatorID != "" {
		req.Header.Set("X-ReleaseOperator-ID", release_operatorID)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, res.Code, wantStatus, res.Body.String())
	}
	if res.Header().Get("X-Request-ID") != "http-contract" {
		t.Fatalf("%s %s request id=%q", method, path, res.Header().Get("X-Request-ID"))
	}
	if target != nil {
		if err := json.Unmarshal(res.Body.Bytes(), target); err != nil {
			t.Fatalf("%s %s decode response: %v body=%s", method, path, err, res.Body.String())
		}
	}
}

func createReleaseOperatorHTTP(t *testing.T, handler http.Handler, name string, role domain.ReleaseOperatorRole) domain.ReleaseOperator {
	t.Helper()
	var release_operator domain.ReleaseOperator
	apiRequest(t, handler, http.MethodPost, "/v1/release_operators", map[string]any{"name": name, "role": role}, "", http.StatusCreated, &release_operator)
	if release_operator.ID == "" || release_operator.Role != role {
		t.Fatalf("release_operator=%+v", release_operator)
	}
	return release_operator
}

func TestHTTPWorkflowCoversPublicBackendRoutes(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	now := time.Now().UTC().Truncate(time.Second)
	svc := service.New(repository.NewPostgres(pool)).WithClock(func() time.Time { return now })
	handler := httpapi.New(svc, pool.Ping).Handler()

	apiRequest(t, handler, http.MethodGet, "/healthz", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/readyz", nil, "", http.StatusOK, nil)

	field := createReleaseOperatorHTTP(t, handler, "ManagedDeviceOperator Lin", domain.RoleManagedDeviceOperator)
	installationOperator := createReleaseOperatorHTTP(t, handler, "InstallationOperator Zhao", domain.RoleInstallationOperator)
	reviewer := createReleaseOperatorHTTP(t, handler, "Safety Reviewer Chen", domain.RoleQualityReviewer)
	apiRequest(t, handler, http.MethodPost, "/v1/release_operators/"+field.ID+"/rename", map[string]any{"name": "Senior ManagedDeviceOperator Lin"}, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/v1/release_operators?role=managed_device_operator", nil, "", http.StatusOK, nil)

	request := map[string]any{
		"code": "HTTP-PLAN", "name": "HTTP workflow", "timezone": "UTC",
		"starts_at": now.Add(-time.Minute), "ends_at": now.Add(24 * time.Hour), "created_by": field.ID,
		"managed_devices": []map[string]any{{"code": "MANAGED_DEVICE-01", "rollout_lane": "A-101", "required_successes": 2}},
	}
	var rollout_campaign service.CreateRolloutCampaignResponse
	apiRequest(t, handler, http.MethodPost, "/v1/cases", request, field.ID, http.StatusCreated, &rollout_campaign)
	if rollout_campaign.RolloutCampaign.ID == "" || len(rollout_campaign.ManagedDeviceIDs) != 1 {
		t.Fatalf("rollout_campaign=%+v", rollout_campaign)
	}
	rollout_campaignID, managed_deviceID := rollout_campaign.RolloutCampaign.ID, rollout_campaign.ManagedDeviceIDs[0]
	apiRequest(t, handler, http.MethodGet, "/v1/cases?search=HTTP", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/v1/cases/"+rollout_campaignID+"/managed_devices", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/cases/"+rollout_campaignID+"/schedule", map[string]any{"version": 1}, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/cases/"+rollout_campaignID+"/activate", map[string]any{"version": 2}, field.ID, http.StatusOK, nil)

	var assignment domain.Assignment
	apiRequest(t, handler, http.MethodPost, "/v1/assignments", map[string]any{
		"rollout_campaign_id": rollout_campaignID, "managed_device_id": managed_deviceID, "release_operator_id": field.ID,
		"starts_at": now.Add(-time.Minute), "ends_at": now.Add(time.Hour),
	}, field.ID, http.StatusCreated, &assignment)
	apiRequest(t, handler, http.MethodPost, "/v1/assignments/"+assignment.ID+"/advance", map[string]any{"status": "active", "version": 1}, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/assignments/"+assignment.ID+"/advance", map[string]any{"status": "completed", "version": 2}, field.ID, http.StatusOK, nil)

	createDeploymentJob := func(code string) domain.DeploymentJob {
		var task domain.DeploymentJob
		apiRequest(t, handler, http.MethodPost, "/v1/deployment_jobs", map[string]any{
			"rollout_campaign_id": rollout_campaignID, "managed_device_id": managed_deviceID, "task_code": code, "expires_at": now.Add(12 * time.Hour),
		}, field.ID, http.StatusCreated, &task)
		apiRequest(t, handler, http.MethodPost, "/v1/deployment_jobs/"+task.ID+"/complete", map[string]any{"version": 1}, field.ID, http.StatusOK, nil)
		apiRequest(t, handler, http.MethodPost, "/v1/deployment_jobs/"+task.ID+"/activation", map[string]any{
			"to_operator": field.ID, "location": "A-101", "recorded_at": now, "version": 2,
		}, field.ID, http.StatusOK, nil)
		apiRequest(t, handler, http.MethodPost, "/v1/deployment_jobs/"+task.ID+"/accept", map[string]any{
			"from_operator": field.ID, "to_operator": installationOperator.ID, "location": "East managed_device station", "recorded_at": now.Add(time.Minute), "version": 3,
		}, installationOperator.ID, http.StatusOK, nil)
		return task
	}

	first := createDeploymentJob("HTTP-TASK-01")
	apiRequest(t, handler, http.MethodGet, "/v1/deployment_jobs?rollout_campaign_id="+rollout_campaignID+"&status=accepted", nil, "", http.StatusOK, nil)
	var rollout_wave map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/rollout_waves", map[string]any{
		"code": "HTTP-ROUND-01", "method": "evening-managed_device", "capacity": 1, "deployment_job_ids": []string{first.ID},
	}, installationOperator.ID, http.StatusCreated, &rollout_wave)
	rollout_waveID := rollout_wave["id"]
	apiRequest(t, handler, http.MethodPost, "/v1/rollout_waves/"+rollout_waveID+"/start?version=1", nil, installationOperator.ID, http.StatusOK, nil)
	var installation_report map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/installation_report", map[string]any{
		"deployment_job_id": first.ID, "rollout_wave_id": rollout_waveID, "recorded_by": installationOperator.ID,
		"risk_score": 3.5, "scale": "managed_device-risk", "alert_threshold": 5.0, "observed_at": now,
	}, installationOperator.ID, http.StatusCreated, &installation_report)
	apiRequest(t, handler, http.MethodPost, "/v1/installation_report/"+installation_report["id"]+"/review", map[string]any{
		"deployment_job_id": first.ID, "accepted": true, "installation_report_version": 1, "task_version": 5,
	}, reviewer.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/rollout_waves/"+rollout_waveID+"/complete?version=2", nil, installationOperator.ID, http.StatusOK, nil)

	var safety_alert map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/health_alerts", map[string]any{
		"deployment_job_id": first.ID, "kind": "close_record", "reason": "scheduled record closure", "due_at": now.Add(time.Hour),
	}, field.ID, http.StatusCreated, &safety_alert)
	apiRequest(t, handler, http.MethodPost, "/v1/health_alerts/"+safety_alert["id"]+"/start", nil, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/health_alerts/"+safety_alert["id"]+"/close", nil, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/deployment_jobs/"+first.ID+"/archive", map[string]any{"version": 6}, field.ID, http.StatusOK, nil)

	second := createDeploymentJob("HTTP-TASK-02")
	var cancelRolloutWave map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/rollout_waves", map[string]any{
		"code": "HTTP-ROUND-02", "method": "evening-managed_device", "capacity": 1, "deployment_job_ids": []string{second.ID},
	}, installationOperator.ID, http.StatusCreated, &cancelRolloutWave)
	apiRequest(t, handler, http.MethodPost, "/v1/rollout_waves/"+cancelRolloutWave["id"]+"/cancel?version=1", nil, installationOperator.ID, http.StatusOK, nil)
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM deployment_jobs WHERE id=$1`, second.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.DeploymentJobAccepted) {
		t.Fatalf("cancelled rollout_wave left task status=%s", status)
	}
	var collected int
	if err := pool.QueryRow(ctx, `SELECT completed_installs FROM managed_devices WHERE id=$1`, managed_deviceID).Scan(&collected); err != nil {
		t.Fatal(err)
	}
	if collected != 2 {
		t.Fatalf("managed_device completed_installs=%d want=2", collected)
	}

	apiRequest(t, handler, http.MethodGet, "/v1/cases/"+rollout_campaignID+"/progress", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/v1/cases/"+rollout_campaignID+"/report", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, fmt.Sprintf("/v1/audit/managed_device_task/%s?limit=20", first.ID), nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/cases/"+rollout_campaignID+"/close", map[string]any{"version": 3}, field.ID, http.StatusOK, nil)
}
