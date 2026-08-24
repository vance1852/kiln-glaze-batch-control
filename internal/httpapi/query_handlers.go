package httpapi

import (
	"net/http"
	"strconv"

	"firmware-rollout-control/internal/domain"
	"firmware-rollout-control/internal/repository"
)

func (a *API) listRolloutCampaigns(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, total, err := a.service.ListRolloutCampaigns(r.Context(), repository.RolloutCampaignFilter{Status: domain.RolloutCampaignStatus(r.URL.Query().Get("status")), Search: r.URL.Query().Get("search"), Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (a *API) listRolloutCampaignManagedDevices(w http.ResponseWriter, r *http.Request) {
	items, err := a.service.ListRolloutCampaignManagedDevices(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (a *API) auditHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.service.AuditHistory(r.Context(), r.PathValue("object_type"), r.PathValue("object_id"), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}
