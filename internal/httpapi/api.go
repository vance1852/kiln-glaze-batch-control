package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"firmware-rollout-control/internal/console"
	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
	"firmware-rollout-control/internal/service"
	"github.com/google/uuid"
)

type API struct {
	service      *service.Service
	ready        func(context.Context) error
	consoleStore *console.Store
}

func New(svc *service.Service, ready func(context.Context) error) *API {
	return &API{service: svc, ready: ready}
}

func (a *API) WithConsole(store *console.Store) *API {
	a.consoleStore = store
	return a
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /readyz", a.readyz)
	mux.HandleFunc("POST /v1/cases", a.createRolloutCampaign)
	mux.HandleFunc("POST /v1/cases/{id}/schedule", a.scheduleRolloutCampaign)
	mux.HandleFunc("POST /v1/cases/{id}/activate", a.activateRolloutCampaign)
	mux.HandleFunc("POST /v1/cases/{id}/close", a.closeRolloutCampaign)
	mux.HandleFunc("GET /v1/cases/{id}/progress", a.rollout_campaignProgress)
	mux.HandleFunc("GET /v1/cases", a.listRolloutCampaigns)
	mux.HandleFunc("GET /v1/cases/{id}/managed_devices", a.listRolloutCampaignManagedDevices)
	mux.HandleFunc("GET /v1/cases/{id}/report", a.complianceReport)
	mux.HandleFunc("POST /v1/release_operators", a.createReleaseOperator)
	mux.HandleFunc("GET /v1/release_operators", a.listReleaseOperators)
	mux.HandleFunc("POST /v1/release_operators/{id}/rename", a.renameReleaseOperator)
	mux.HandleFunc("POST /v1/assignments", a.createAssignment)
	mux.HandleFunc("POST /v1/assignments/{id}/advance", a.advanceAssignment)
	mux.HandleFunc("POST /v1/deployment_jobs", a.createDeploymentJob)
	mux.HandleFunc("POST /v1/deployment_jobs/{id}/complete", a.completeDeploymentJob)
	mux.HandleFunc("POST /v1/deployment_jobs/{id}/activation", a.transferDeploymentJob)
	mux.HandleFunc("POST /v1/deployment_jobs/{id}/accept", a.acceptDeploymentJob)
	mux.HandleFunc("POST /v1/deployment_jobs/{id}/archive", a.archiveDeploymentJob)
	mux.HandleFunc("GET /v1/deployment_jobs", a.listDeploymentJobs)
	mux.HandleFunc("POST /v1/rollout_waves", a.createRolloutWave)
	mux.HandleFunc("POST /v1/rollout_waves/{id}/start", a.startRolloutWave)
	mux.HandleFunc("POST /v1/rollout_waves/{id}/complete", a.completeRolloutWave)
	mux.HandleFunc("POST /v1/rollout_waves/{id}/cancel", a.cancelRolloutWave)
	mux.HandleFunc("POST /v1/installation_report", a.submitInstallationReport)
	mux.HandleFunc("POST /v1/installation_report/{id}/review", a.reviewInstallationReport)
	mux.HandleFunc("POST /v1/health_alerts", a.openHealthAlert)
	mux.HandleFunc("POST /v1/health_alerts/{id}/start", a.startHealthAlert)
	mux.HandleFunc("POST /v1/health_alerts/{id}/close", a.closeHealthAlert)
	mux.HandleFunc("GET /v1/audit", a.queryAudit)
	mux.HandleFunc("GET /v1/audit/{object_type}/{object_id}", a.auditHistory)
	if a.consoleStore != nil {
		a.registerConsoleRoutes(mux)
	}
	return requestMiddleware(mux)
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (a *API) readyz(w http.ResponseWriter, r *http.Request) {
	if a.ready != nil {
		if err := a.ready(r.Context()); err != nil {
			writeError(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type rollout_campaignRequest struct {
	Code           string                          `json:"code"`
	Name           string                          `json:"name"`
	Timezone       string                          `json:"timezone"`
	StartsAt       time.Time                       `json:"starts_at"`
	EndsAt         time.Time                       `json:"ends_at"`
	CreatedBy      string                          `json:"created_by"`
	ManagedDevices []repository.ManagedDeviceInput `json:"managed_devices"`
}

func (a *API) createRolloutCampaign(w http.ResponseWriter, r *http.Request) {
	var in rollout_campaignRequest
	if !decode(w, r, &in) {
		return
	}
	result, err := a.service.CreateRolloutCampaign(r.Context(), meta(r), service.CreateRolloutCampaignRequest{Code: in.Code, Name: in.Name, Timezone: in.Timezone, StartsAt: in.StartsAt, EndsAt: in.EndsAt, CreatedBy: in.CreatedBy, ManagedDevices: in.ManagedDevices})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

type versionRequest struct {
	Version int64 `json:"version"`
}

func (a *API) scheduleRolloutCampaign(w http.ResponseWriter, r *http.Request) {
	a.advanceRolloutCampaign(w, r, domain.RolloutCampaignScheduled)
}
func (a *API) activateRolloutCampaign(w http.ResponseWriter, r *http.Request) {
	a.advanceRolloutCampaign(w, r, domain.RolloutCampaignActive)
}
func (a *API) closeRolloutCampaign(w http.ResponseWriter, r *http.Request) {
	a.advanceRolloutCampaign(w, r, domain.RolloutCampaignClosed)
}

func (a *API) advanceRolloutCampaign(w http.ResponseWriter, r *http.Request, next domain.RolloutCampaignStatus) {
	var in versionRequest
	if !decode(w, r, &in) {
		return
	}
	var err error
	switch next {
	case domain.RolloutCampaignScheduled:
		err = a.service.ScheduleRolloutCampaign(r.Context(), meta(r), r.PathValue("id"), in.Version)
	case domain.RolloutCampaignActive:
		err = a.service.ActivateRolloutCampaign(r.Context(), meta(r), r.PathValue("id"), in.Version)
	case domain.RolloutCampaignClosed:
		err = a.service.CloseRolloutCampaign(r.Context(), meta(r), r.PathValue("id"), in.Version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(next)})
}

func (a *API) createDeploymentJob(w http.ResponseWriter, r *http.Request) {
	var in repository.DeploymentJobInput
	if !decode(w, r, &in) {
		return
	}
	result, err := a.service.CreateDeploymentJob(r.Context(), meta(r), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (a *API) completeDeploymentJob(w http.ResponseWriter, r *http.Request) {
	a.moveDeploymentJob(w, r, domain.DeploymentJobCompleted)
}

func (a *API) archiveDeploymentJob(w http.ResponseWriter, r *http.Request) {
	a.moveDeploymentJob(w, r, domain.DeploymentJobArchived)
}

func (a *API) moveDeploymentJob(w http.ResponseWriter, r *http.Request, next domain.DeploymentJobStatus) {
	var in versionRequest
	if !decode(w, r, &in) {
		return
	}
	var err error
	if next == domain.DeploymentJobCompleted {
		err = a.service.CompleteDeploymentJob(r.Context(), meta(r), r.PathValue("id"), in.Version)
	} else {
		err = a.service.ArchiveDeploymentJob(r.Context(), meta(r), r.PathValue("id"), in.Version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(next)})
}

func (a *API) transferDeploymentJob(w http.ResponseWriter, r *http.Request) {
	a.activation(w, r, domain.DeploymentJobActivationPending)
}

func (a *API) acceptDeploymentJob(w http.ResponseWriter, r *http.Request) {
	a.activation(w, r, domain.DeploymentJobAccepted)
}

type activationRequest struct {
	From       *string   `json:"from_operator"`
	To         string    `json:"to_operator"`
	Location   string    `json:"location"`
	RecordedAt time.Time `json:"recorded_at"`
	Note       string    `json:"note"`
	Version    int64     `json:"version"`
}

func (a *API) activation(w http.ResponseWriter, r *http.Request, next domain.DeploymentJobStatus) {
	var in activationRequest
	if !decode(w, r, &in) {
		return
	}
	if in.RecordedAt.IsZero() {
		in.RecordedAt = time.Now().UTC()
	}
	input := repository.ActivationInput{DeploymentJobID: r.PathValue("id"), From: in.From, To: in.To, Location: in.Location, RecordedAt: in.RecordedAt, Note: in.Note}
	var err error
	if next == domain.DeploymentJobActivationPending {
		err = a.service.ActivationChecked(r.Context(), meta(r), input, in.Version)
	} else {
		err = a.service.AcceptChecked(r.Context(), meta(r), input, in.Version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(next)})
}

func (a *API) listDeploymentJobs(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	page, err := a.service.ListDeploymentJobs(r.Context(), offset, limit, r.URL.Query().Get("rollout_campaign_id"), domain.DeploymentJobStatus(r.URL.Query().Get("status")))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

type rollout_waveRequest struct {
	repository.RolloutWaveInput
	DeploymentJobIDs []string `json:"deployment_job_ids"`
}

func (a *API) createRolloutWave(w http.ResponseWriter, r *http.Request) {
	var in rollout_waveRequest
	if !decode(w, r, &in) {
		return
	}
	id, err := a.service.CreateRolloutWave(r.Context(), meta(r), in.RolloutWaveInput, append([]string(nil), in.DeploymentJobIDs...))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (a *API) submitInstallationReport(w http.ResponseWriter, r *http.Request) {
	var in repository.InstallationReportInput
	if !decode(w, r, &in) {
		return
	}
	id, err := a.service.SubmitInstallationReport(r.Context(), meta(r), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

type reviewRequest struct {
	DeploymentJobID           string `json:"deployment_job_id"`
	Accepted                  bool   `json:"accepted"`
	InstallationReportVersion int64  `json:"installation_report_version"`
	DeploymentJobVersion      int64  `json:"task_version"`
}

func (a *API) reviewInstallationReport(w http.ResponseWriter, r *http.Request) {
	var in reviewRequest
	if !decode(w, r, &in) {
		return
	}
	err := a.service.ReviewInstallationReport(r.Context(), meta(r), r.PathValue("id"), in.DeploymentJobID, in.Accepted, in.InstallationReportVersion, in.DeploymentJobVersion)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reviewed"})
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, fmt.Errorf("invalid json: %w", domain.ErrConflict))
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, fmt.Errorf("request body must contain one json value: %w", domain.ErrConflict))
		return false
	}
	return true
}

func meta(r *http.Request) service.RequestMeta {
	release_operator := strings.TrimSpace(r.Header.Get("X-ReleaseOperator-ID"))
	var release_operatorID *string
	if _, err := uuid.Parse(release_operator); err == nil {
		release_operatorID = &release_operator
	}
	return service.RequestMeta{RequestID: requestID(r.Context()), ReleaseOperatorID: release_operatorID}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrInvalidTransition):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, domain.ErrCapacityExceeded):
		status, code = http.StatusUnprocessableEntity, "capacity_exceeded"
	case errors.Is(err, domain.ErrExpired):
		status, code = http.StatusUnprocessableEntity, "expired"
	}
	writeJSON(w, status, map[string]string{"code": code, "message": err.Error(), "request_id": requestIDFromWriter(w)})
}

func requestIDFromWriter(w http.ResponseWriter) string { return w.Header().Get("X-Request-ID") }
