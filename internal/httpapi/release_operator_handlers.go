package httpapi

import (
	"net/http"

	"firmware-rollout-control/internal/domain"
)

func (a *API) listReleaseOperators(w http.ResponseWriter, r *http.Request) {
	items, total, err := a.service.ListReleaseOperators(r.Context(), domain.ReleaseOperatorRole(r.URL.Query().Get("role")), 50, 0)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

type renameReleaseOperatorRequest struct {
	Name string `json:"name"`
}

func (a *API) renameReleaseOperator(w http.ResponseWriter, r *http.Request) {
	var in renameReleaseOperatorRequest
	if !decode(w, r, &in) {
		return
	}
	if err := a.service.RenameReleaseOperator(r.Context(), meta(r), r.PathValue("id"), in.Name); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "renamed"})
}
