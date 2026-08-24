package httpapi

import (
	"net/http"

	"firmware-rollout-control/internal/repository"
)

func (a *API) openHealthAlert(w http.ResponseWriter, r *http.Request) {
	var in repository.HealthAlertInput
	if !decode(w, r, &in) {
		return
	}
	id, err := a.service.OpenHealthAlert(r.Context(), meta(r), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (a *API) closeHealthAlert(w http.ResponseWriter, r *http.Request) {
	if err := a.service.CloseHealthAlert(r.Context(), meta(r), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func (a *API) startHealthAlert(w http.ResponseWriter, r *http.Request) {
	if err := a.service.MarkHealthAlertInProgress(r.Context(), meta(r), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "in_progress"})
}
