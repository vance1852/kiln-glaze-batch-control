package console

import "time"

type Role string

const (
	RoleCurator         Role = "curator"
	RoleReleaseOperator Role = "release_operator"
	RoleReviewer        Role = "reviewer"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Role     Role   `json:"role"`
	Status   int    `json:"status"`
}

type Session struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ManagedDevice struct {
	ID                      string  `json:"id"`
	Title                   string  `json:"title"`
	DeviceClass             int     `json:"materialClass"`
	Comrollout_campaignedOn *string `json:"catalogedOn"`
	AccessionNumber         string  `json:"accessionNumber"`
	RepositoryCode          string  `json:"repositoryCode"`
	StorageZone             string  `json:"storageZone"`
	DonorName               string  `json:"donorName"`
	CuratorContact          string  `json:"curatorContact"`
	ConditionStatus         int     `json:"conditionStatus"`
	Status                  int     `json:"status"`
}

type ReleaseOperator struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	SpecialtyLevel int       `json:"specialtyLevel"`
	Phone          string    `json:"phone"`
	Skills         string    `json:"skills"`
	Status         int       `json:"status"`
	CreateTime     time.Time `json:"createTime"`
}

type Treatment struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	RiskBudget      float64 `json:"riskBudget"`
	DurationMinutes int     `json:"durationMinutes"`
	Status          int     `json:"status"`
}

type WorkOrder struct {
	ID                  string     `json:"id"`
	WorkOrderNo         string     `json:"workOrderNo"`
	ManagedDeviceID     string     `json:"managed_deviceId"`
	ManagedDeviceTitle  string     `json:"managed_deviceTitle"`
	RolloutProfileID    string     `json:"treatmentId"`
	TreatmentName       string     `json:"treatmentName"`
	ReleaseOperatorID   *string    `json:"release_operatorId"`
	ReleaseOperatorName string     `json:"release_operatorName"`
	ScheduledAt         *time.Time `json:"scheduledAt"`
	Status              int        `json:"status"`
	Remark              string     `json:"remark"`
	Version             int64      `json:"version"`
}

type InstallationReport struct {
	ID                 string    `json:"id"`
	ManagedDeviceID    string    `json:"managed_deviceId"`
	ManagedDeviceTitle string    `json:"managed_deviceTitle"`
	RelativeHumidity   *float64  `json:"relativeHumidity"`
	TemperatureC       *float64  `json:"temperatureC"`
	IlluminanceLux     *float64  `json:"illuminanceLux"`
	AcidityPH          *float64  `json:"acidityPh"`
	PestIndex          *float64  `json:"pestIndex"`
	Remark             string    `json:"remark"`
	RecordedAt         time.Time `json:"recordedAt"`
}

type Log struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Operation  string    `json:"operation"`
	Method     string    `json:"method"`
	IP         string    `json:"ip"`
	CreateTime time.Time `json:"createTime"`
}

type Page[T any] struct {
	Records []T `json:"records"`
	Total   int `json:"total"`
	Current int `json:"current"`
	Size    int `json:"size"`
}

type DashboardStats struct {
	ManagedDeviceCount   int `json:"managed_deviceCount"`
	ReleaseOperatorCount int `json:"release_operatorCount"`
	PendingWorkOrders    int `json:"pendingWorkOrders"`
	CompletedWorkOrders  int `json:"completedWorkOrders"`
}
