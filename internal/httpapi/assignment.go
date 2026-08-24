package httpapi

import (
	"net/http"
	"time"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

type assignmentRequest struct {
	RolloutCampaignID string    `json:"rollout_campaign_id"`
	ManagedDeviceID   string    `json:"managed_device_id"`
	ReleaseOperatorID string    `json:"release_operator_id"`
	StartsAt          time.Time `json:"starts_at"`
	EndsAt            time.Time `json:"ends_at"`
}

func (a *API) createAssignment(w http.ResponseWriter, r *http.Request) {
	var in assignmentRequest
	if !decode(w, r, &in) {
		return
	}
	release_operator, err := a.service.LoadReleaseOperator(r.Context(), in.ReleaseOperatorID)
	if err != nil {
		writeError(w, err)
		return
	}
	assignment := repository.NewAssignment(in.RolloutCampaignID, in.ManagedDeviceID, in.ReleaseOperatorID, in.StartsAt, in.EndsAt)
	if err := a.service.AssignManagedDevice(r.Context(), meta(r), assignment, release_operator); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, assignment)
}

type assignmentAdvanceRequest struct {
	Status  string `json:"status"`
	Version int64  `json:"version"`
}

func (a *API) advanceAssignment(w http.ResponseWriter, r *http.Request) {
	var in assignmentAdvanceRequest
	if !decode(w, r, &in) {
		return
	}
	if in.Status != "active" && in.Status != "completed" && in.Status != "cancelled" {
		writeError(w, domain.ErrConflict)
		return
	}
	if err := a.service.AdvanceAssignment(r.Context(), meta(r), r.PathValue("id"), in.Status, in.Version); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": in.Status})
}
