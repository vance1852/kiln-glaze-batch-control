package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"firmware-rollout-control/internal/console"
)

func (a *API) registerConsoleRoutes(mux *http.ServeMux) {
	allRoles := []console.Role{console.RoleCurator, console.RoleReleaseOperator, console.RoleReviewer}
	curator := []console.Role{console.RoleCurator}
	curatorAndReviewer := []console.Role{console.RoleCurator, console.RoleReviewer}
	curatorAndReleaseOperator := []console.Role{console.RoleCurator, console.RoleReleaseOperator}
	mux.HandleFunc("POST /api/auth/login", a.consoleLogin)
	mux.HandleFunc("GET /api/auth/info", a.requireConsoleRole(a.consoleUserInfo, allRoles...))
	mux.HandleFunc("POST /api/auth/logout", a.requireConsoleRole(a.consoleLogout, allRoles...))
	mux.HandleFunc("GET /api/dashboard/stats", a.requireConsoleRole(a.consoleDashboard, allRoles...))
	mux.HandleFunc("GET /api/managed_device/page", a.requireConsoleRole(a.consoleManagedDevicePage, allRoles...))
	mux.HandleFunc("GET /api/managed_device/list", a.requireConsoleRole(a.consoleManagedDeviceList, allRoles...))
	mux.HandleFunc("GET /api/managed_device/{id}", a.requireConsoleRole(a.consoleManagedDeviceByID, allRoles...))
	mux.HandleFunc("POST /api/managed_device", a.requireConsoleRole(a.consoleCreateManagedDevice, curator...))
	mux.HandleFunc("PUT /api/managed_device", a.requireConsoleRole(a.consoleUpdateManagedDevice, curator...))
	mux.HandleFunc("DELETE /api/managed_device/{id}", a.requireConsoleRole(a.consoleDeleteManagedDevice, curator...))
	mux.HandleFunc("GET /api/release_operator/page", a.requireConsoleRole(a.consoleReleaseOperatorPage, allRoles...))
	mux.HandleFunc("GET /api/release_operator/list", a.requireConsoleRole(a.consoleReleaseOperatorList, allRoles...))
	mux.HandleFunc("POST /api/release_operator", a.requireConsoleRole(a.consoleCreateReleaseOperator, curator...))
	mux.HandleFunc("PUT /api/release_operator", a.requireConsoleRole(a.consoleUpdateReleaseOperator, curator...))
	mux.HandleFunc("DELETE /api/release_operator/{id}", a.requireConsoleRole(a.consoleDeleteReleaseOperator, curator...))
	mux.HandleFunc("GET /api/treatment/page", a.requireConsoleRole(a.consoleTreatmentPage, allRoles...))
	mux.HandleFunc("GET /api/treatment/list", a.requireConsoleRole(a.consoleTreatmentList, allRoles...))
	mux.HandleFunc("POST /api/treatment", a.requireConsoleRole(a.consoleCreateTreatment, curatorAndReviewer...))
	mux.HandleFunc("PUT /api/treatment", a.requireConsoleRole(a.consoleUpdateTreatment, curatorAndReviewer...))
	mux.HandleFunc("DELETE /api/treatment/{id}", a.requireConsoleRole(a.consoleDeleteTreatment, curator...))
	mux.HandleFunc("GET /api/work_order/page", a.requireConsoleRole(a.consoleWorkOrderPage, allRoles...))
	mux.HandleFunc("POST /api/work_order", a.requireConsoleRole(a.consoleCreateWorkOrder, curatorAndReleaseOperator...))
	mux.HandleFunc("PUT /api/work_order/status", a.requireConsoleRole(a.consoleUpdateWorkOrderStatus, allRoles...))
	mux.HandleFunc("GET /api/installation_report/page", a.requireConsoleRole(a.consoleInstallationReportPage, allRoles...))
	mux.HandleFunc("GET /api/installation_report/managed_device/{managed_deviceId}", a.requireConsoleRole(a.consoleInstallationReportByManagedDevice, allRoles...))
	mux.HandleFunc("POST /api/installation_report", a.requireConsoleRole(a.consoleCreateInstallationReport, curatorAndReleaseOperator...))
	mux.HandleFunc("GET /api/log/page", a.requireConsoleRole(a.consoleLogPage, curatorAndReviewer...))
}

type consoleAuthContextKey struct{}

type consoleAuth struct {
	User  console.User
	Token string
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(value, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(value, prefix))
	}
	return strings.TrimSpace(r.Header.Get("X-Session-Token"))
}

func (a *API) requireConsoleRole(next http.HandlerFunc, allowed ...console.Role) http.HandlerFunc {
	roles := make(map[console.Role]struct{}, len(allowed))
	for _, role := range allowed {
		roles[role] = struct{}{}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			consoleError(w, http.StatusUnauthorized, fmt.Errorf("缺少登录会话"))
			return
		}
		user, err := a.consoleStore.UserBySession(r.Context(), token)
		if err != nil {
			consoleError(w, http.StatusUnauthorized, err)
			return
		}
		if _, ok := roles[user.Role]; !ok {
			consoleError(w, http.StatusForbidden, fmt.Errorf("当前角色无权执行该操作"))
			return
		}
		ctx := context.WithValue(r.Context(), consoleAuthContextKey{}, consoleAuth{User: user, Token: token})
		next(w, r.WithContext(ctx))
	}
}

func consoleIdentity(r *http.Request) consoleAuth {
	identity, _ := r.Context().Value(consoleAuthContextKey{}).(consoleAuth)
	return identity
}

type consoleEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func consoleSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(consoleEnvelope{Code: 200, Message: "success", Data: data})
}

func consoleError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(consoleEnvelope{Code: status, Message: err.Error()})
}

func decodeConsole(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		consoleError(w, http.StatusBadRequest, fmt.Errorf("请求数据格式错误"))
		return false
	}
	return true
}

func consolePageParams(r *http.Request) (int, int) {
	current, _ := strconv.Atoi(r.URL.Query().Get("current"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	return current, size
}

func consoleClientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, found := strings.Cut(r.RemoteAddr, ":")
	if found {
		return host
	}
	return r.RemoteAddr
}

func (a *API) recordConsoleOperation(r *http.Request, operation string) {
	username := consoleIdentity(r).User.Username
	if username == "" {
		username = "anonymous"
	}
	_ = a.consoleStore.WriteLog(r.Context(), username, operation, r.Method+" "+r.URL.Path, consoleClientIP(r))
}

func (a *API) consoleLogin(w http.ResponseWriter, r *http.Request) {
	var request struct{ Username, Password string }
	if !decodeConsole(w, r, &request) {
		return
	}
	session, user, err := a.consoleStore.Login(r.Context(), request.Username, request.Password, 8*time.Hour)
	if err != nil {
		consoleError(w, http.StatusUnauthorized, err)
		return
	}
	_ = a.consoleStore.WriteLog(r.Context(), user.Username, "用户登录", r.Method+" "+r.URL.Path, consoleClientIP(r))
	consoleSuccess(w, map[string]any{"token": session.Token, "expiresAt": session.ExpiresAt, "user": user})
}

func (a *API) consoleUserInfo(w http.ResponseWriter, r *http.Request) {
	consoleSuccess(w, consoleIdentity(r).User)
}

func (a *API) consoleLogout(w http.ResponseWriter, r *http.Request) {
	identity := consoleIdentity(r)
	if err := a.consoleStore.RevokeSession(r.Context(), identity.Token); err != nil {
		consoleError(w, http.StatusUnauthorized, err)
		return
	}
	consoleSuccess(w, nil)
}

func (a *API) consoleDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := a.consoleStore.Dashboard(r.Context())
	if err != nil {
		consoleError(w, http.StatusInternalServerError, err)
		return
	}
	consoleSuccess(w, stats)
}

func (a *API) consoleManagedDevicePage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.ManagedDevicePage(r.Context(), current, size, r.URL.Query().Get("keyword"))
	consoleWriteResult(w, page, err)
}

func (a *API) consoleManagedDeviceList(w http.ResponseWriter, r *http.Request) {
	items, err := a.consoleStore.ManagedDeviceList(r.Context())
	consoleWriteResult(w, items, err)
}

func (a *API) consoleManagedDeviceByID(w http.ResponseWriter, r *http.Request) {
	item, err := a.consoleStore.ManagedDeviceByID(r.Context(), r.PathValue("id"))
	consoleWriteResult(w, item, err)
}

func (a *API) consoleCreateManagedDevice(w http.ResponseWriter, r *http.Request) {
	var item console.ManagedDevice
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Title) {
		return
	}
	created, err := a.consoleStore.CreateManagedDevice(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增馆藏芯片批次")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateManagedDevice(w http.ResponseWriter, r *http.Request) {
	var item console.ManagedDevice
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Title) {
		return
	}
	err := a.consoleStore.UpdateManagedDevice(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "编辑馆藏芯片批次")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleDeleteManagedDevice(w http.ResponseWriter, r *http.Request) {
	err := a.consoleStore.DeleteManagedDevice(r.Context(), r.PathValue("id"))
	if err == nil {
		a.recordConsoleOperation(r, "删除馆藏芯片批次")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleReleaseOperatorPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.ReleaseOperatorPage(r.Context(), current, size, r.URL.Query().Get("keyword"))
	consoleWriteResult(w, page, err)
}

func (a *API) consoleReleaseOperatorList(w http.ResponseWriter, r *http.Request) {
	items, err := a.consoleStore.ReleaseOperatorList(r.Context())
	consoleWriteResult(w, items, err)
}

func (a *API) consoleCreateReleaseOperator(w http.ResponseWriter, r *http.Request) {
	var item console.ReleaseOperator
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	created, err := a.consoleStore.CreateReleaseOperator(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增校准操作员")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateReleaseOperator(w http.ResponseWriter, r *http.Request) {
	var item console.ReleaseOperator
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	err := a.consoleStore.UpdateReleaseOperator(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "编辑校准操作员")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleDeleteReleaseOperator(w http.ResponseWriter, r *http.Request) {
	err := a.consoleStore.DeleteReleaseOperator(r.Context(), r.PathValue("id"))
	if err == nil {
		a.recordConsoleOperation(r, "删除校准操作员")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleTreatmentPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.TreatmentPage(r.Context(), current, size, r.URL.Query().Get("keyword"))
	consoleWriteResult(w, page, err)
}

func (a *API) consoleTreatmentList(w http.ResponseWriter, r *http.Request) {
	items, err := a.consoleStore.TreatmentList(r.Context())
	consoleWriteResult(w, items, err)
}

func (a *API) consoleCreateTreatment(w http.ResponseWriter, r *http.Request) {
	var item console.Treatment
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	created, err := a.consoleStore.CreateTreatment(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增修复方案")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateTreatment(w http.ResponseWriter, r *http.Request) {
	var item console.Treatment
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	err := a.consoleStore.UpdateTreatment(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "编辑修复方案")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleDeleteTreatment(w http.ResponseWriter, r *http.Request) {
	err := a.consoleStore.DeleteTreatment(r.Context(), r.PathValue("id"))
	if err == nil {
		a.recordConsoleOperation(r, "删除修复方案")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleWorkOrderPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.WorkOrderPage(r.Context(), current, size)
	consoleWriteResult(w, page, err)
}

type consoleWorkOrderRequest struct {
	ManagedDeviceID   string  `json:"managed_deviceId"`
	RolloutProfileID  string  `json:"treatmentId"`
	ReleaseOperatorID *string `json:"release_operatorId"`
	ScheduledAt       string  `json:"scheduledAt"`
	Remark            string  `json:"remark"`
}

func (a *API) consoleCreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	var request consoleWorkOrderRequest
	if !decodeConsole(w, r, &request) {
		return
	}
	appointment, err := parseConsoleTime(request.ScheduledAt)
	if err != nil {
		consoleError(w, http.StatusBadRequest, err)
		return
	}
	created, err := a.consoleStore.CreateWorkOrder(r.Context(), console.WorkOrder{ManagedDeviceID: request.ManagedDeviceID, RolloutProfileID: request.RolloutProfileID, ReleaseOperatorID: request.ReleaseOperatorID, ScheduledAt: appointment, Remark: request.Remark})
	if err == nil {
		a.recordConsoleOperation(r, "创建修复工单")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateWorkOrderStatus(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     string `json:"id"`
		Status int    `json:"status"`
	}
	if !decodeConsole(w, r, &request) {
		return
	}
	err := a.consoleStore.UpdateWorkOrderStatus(r.Context(), request.ID, request.Status)
	if err == nil {
		a.recordConsoleOperation(r, "更新修复工单状态")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleInstallationReportPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.InstallationReportPage(r.Context(), current, size, "")
	consoleWriteResult(w, page, err)
}

func (a *API) consoleInstallationReportByManagedDevice(w http.ResponseWriter, r *http.Request) {
	page, err := a.consoleStore.InstallationReportPage(r.Context(), 1, 100, r.PathValue("managed_deviceId"))
	consoleWriteResult(w, page.Records, err)
}

func (a *API) consoleCreateInstallationReport(w http.ResponseWriter, r *http.Request) {
	var item console.InstallationReport
	if !decodeConsole(w, r, &item) {
		return
	}
	created, err := a.consoleStore.CreateInstallationReport(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增舱室环境记录")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleLogPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.LogPage(r.Context(), current, size)
	consoleWriteResult(w, page, err)
}

func consoleWriteResult(w http.ResponseWriter, data any, err error) {
	if err != nil {
		consoleError(w, http.StatusConflict, err)
		return
	}
	consoleSuccess(w, data)
}

func validateConsoleName(w http.ResponseWriter, name string) bool {
	if strings.TrimSpace(name) == "" {
		consoleError(w, http.StatusBadRequest, fmt.Errorf("名称不能为空"))
		return false
	}
	return true
}

func parseConsoleTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			utc := parsed.UTC()
			return &utc, nil
		}
	}
	return nil, fmt.Errorf("计划处理时间格式错误")
}
