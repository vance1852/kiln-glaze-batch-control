package httpapi

import (
	"net/http"
	"strconv"

	"firmware-rollout-control/internal/domain"
)

type release_operatorRequest struct {
	Name string                     `json:"name"`
	Role domain.ReleaseOperatorRole `json:"role"`
}

func (a *API) createReleaseOperator(w http.ResponseWriter, r *http.Request) {
	var in release_operatorRequest
	if !decode(w, r, &in) {
		return
	}
	release_operator, err := a.service.RegisterReleaseOperator(r.Context(), meta(r), in.Name, in.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, release_operator)
}

func (a *API) rollout_campaignProgress(w http.ResponseWriter, r *http.Request) {
	progress, err := a.service.RolloutCampaignProgress(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

func (a *API) rollout_waveVersion(w http.ResponseWriter, r *http.Request) (int64, bool) {
	version, err := strconv.ParseInt(r.URL.Query().Get("version"), 10, 64)
	if err != nil || version < 1 {
		writeError(w, domain.ErrConflict)
		return 0, false
	}
	return version, true
}

func (a *API) startRolloutWave(w http.ResponseWriter, r *http.Request) {
	a.changeRolloutWave(w, r, "start")
}
func (a *API) completeRolloutWave(w http.ResponseWriter, r *http.Request) {
	a.changeRolloutWave(w, r, "complete")
}
func (a *API) cancelRolloutWave(w http.ResponseWriter, r *http.Request) {
	a.changeRolloutWave(w, r, "cancel")
}

func (a *API) changeRolloutWave(w http.ResponseWriter, r *http.Request, action string) {
	version, ok := a.rollout_waveVersion(w, r)
	if !ok {
		return
	}
	var err error
	switch action {
	case "start":
		err = a.service.StartRolloutWave(r.Context(), meta(r), r.PathValue("id"), version)
	case "complete":
		err = a.service.CompleteRolloutWave(r.Context(), meta(r), r.PathValue("id"), version)
	case "cancel":
		err = a.service.CancelRolloutWave(r.Context(), meta(r), r.PathValue("id"), version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": action})
}
