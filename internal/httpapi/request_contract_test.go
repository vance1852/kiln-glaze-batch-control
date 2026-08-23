package httpapi

import (
	"encoding/json"
	"testing"
	"time"

	"firmware-rollout-control/internal/repository"
)

func TestRepositoryInputsDecodePublicJSONFields(t *testing.T) {
	observedAt := "2026-08-18T09:00:00Z"

	var task repository.DeploymentJobInput
	decodeContractJSON(t, `{"rollout_campaign_id":"rollout_campaign-1","managed_device_id":"managed_device-1","task_code":"S-1","expires_at":"2026-08-18T10:00:00Z"}`, &task)
	if task.RolloutCampaignID != "rollout_campaign-1" || task.ManagedDeviceID != "managed_device-1" || task.TaskCode != "S-1" || task.ExpiresAt.IsZero() {
		t.Fatalf("task input = %+v", task)
	}

	var rollout_wave rollout_waveRequest
	decodeContractJSON(t, `{"code":"ROUND-1","method":"evening-managed_device","capacity":2,"deployment_job_ids":["task-1"]}`, &rollout_wave)
	if rollout_wave.Code != "ROUND-1" || rollout_wave.Method != "evening-managed_device" || rollout_wave.Capacity != 2 || len(rollout_wave.DeploymentJobIDs) != 1 {
		t.Fatalf("rollout_wave input = %+v", rollout_wave)
	}

	var installation_report repository.InstallationReportInput
	decodeContractJSON(t, `{"deployment_job_id":"task-1","rollout_wave_id":"round-1","recorded_by":"installationOperator-1","risk_score":2.5,"scale":"managed_device-risk","alert_threshold":5,"observed_at":"`+observedAt+`"}`, &installation_report)
	if installation_report.DeploymentJobID != "task-1" || installation_report.RolloutWaveID != "round-1" || installation_report.RecorderID != "installationOperator-1" || installation_report.ObservedAt.IsZero() {
		t.Fatalf("installation_report input = %+v", installation_report)
	}

	var safety_alert repository.HealthAlertInput
	decodeContractJSON(t, `{"deployment_job_id":"task-1","kind":"repeat_managed_device","reason":"verification","due_at":"2026-08-18T11:00:00Z"}`, &safety_alert)
	if safety_alert.DeploymentJobID != "task-1" || safety_alert.Kind != "repeat_managed_device" || safety_alert.DueAt.IsZero() {
		t.Fatalf("safety_alert input = %+v", safety_alert)
	}

	if installation_report.ObservedAt.Format(time.RFC3339) != observedAt {
		t.Fatalf("observed_at = %s", installation_report.ObservedAt.Format(time.RFC3339))
	}
}

func decodeContractJSON(t *testing.T, value string, destination any) {
	t.Helper()
	if err := json.Unmarshal([]byte(value), destination); err != nil {
		t.Fatal(err)
	}
}
