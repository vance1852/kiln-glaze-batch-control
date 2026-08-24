package rollout

import (
	"errors"
	"time"
)

var (
	ErrConflict     = errors.New("rollout conflict")
	ErrUnauthorized = errors.New("rollout unauthorized")
	ErrUnavailable  = errors.New("rollout unavailable")
	ErrInvalid      = errors.New("rollout invalid")
)

type Artifact struct {
	ID            string
	TenantID      string
	Version       string
	Digest        string
	DeviceClasses []string
	Labels        map[string]string
	Signed        bool
}

type Device struct {
	ID              string
	TenantID        string
	Class           string
	CurrentVersion  string
	PreviousVersion string
	Generation      int64
	Quarantined     bool
}

type Campaign struct {
	ID              string
	TenantID        string
	ArtifactID      string
	State           string
	RequiredHealthy int
	Healthy         int
	Failed          int
	DeviceIDs       []string
	ApprovedDigest  string
	Version         int64
}

type Callback struct {
	TenantID   string
	DeviceID   string
	ArtifactID string
	EventID    string
	Status     string
	At         time.Time
}

type Session struct {
	Token     string
	TenantID  string
	UserID    string
	Role      string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type Event struct {
	TenantID  string
	RequestID string
	ObjectID  string
	Action    string
	Digest    string
	At        time.Time
}

type Query struct {
	TenantID string
	State    string
	Class    string
	Limit    int
	Offset   int
}

type Page[T any] struct {
	Items []T
	Total int
}
