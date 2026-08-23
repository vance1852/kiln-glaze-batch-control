package rollout

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

type Store struct {
	mu          sync.Mutex
	artifacts   map[string]Artifact
	devices     map[string]Device
	campaigns   map[string]Campaign
	callbacks   map[string]Callback
	idempotency map[string][]byte
	sessions    map[string]Session
	events      []Event
}

func NewStore() *Store {
	return &Store{
		artifacts:   make(map[string]Artifact),
		devices:     make(map[string]Device),
		campaigns:   make(map[string]Campaign),
		callbacks:   make(map[string]Callback),
		idempotency: make(map[string][]byte),
		sessions:    make(map[string]Session),
	}
}

func (s *Store) PutArtifact(artifact Artifact) error {
	if artifact.ID == "" || artifact.TenantID == "" {
		return fmt.Errorf("artifact identity is missing: %w", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.artifacts[artifact.ID] = CloneArtifact(artifact)
	return nil
}

func (s *Store) Artifact(id string) (Artifact, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	artifact, ok := s.artifacts[id]
	return CloneArtifact(artifact), ok
}

func (s *Store) PutDevice(device Device) error {
	if device.ID == "" || device.TenantID == "" {
		return fmt.Errorf("device identity is missing: %w", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[device.ID] = device
	return nil
}

func (s *Store) Device(id string) (Device, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	device, ok := s.devices[id]
	return device, ok
}

func (s *Store) PutCampaign(campaign Campaign) error {
	if campaign.ID == "" || campaign.TenantID == "" {
		return fmt.Errorf("campaign identity is missing: %w", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.campaigns[campaign.ID] = CloneCampaign(campaign)
	return nil
}

func (s *Store) Campaign(id string) (Campaign, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	campaign, ok := s.campaigns[id]
	return CloneCampaign(campaign), ok
}

func (s *Store) UpdateCampaign(id string, expected int64, update func(*Campaign) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	campaign, ok := s.campaigns[id]
	if !ok {
		return fmt.Errorf("campaign not found: %w", ErrConflict)
	}
	if campaign.Version != expected {
		return fmt.Errorf("campaign version changed: %w", ErrConflict)
	}
	if err := update(&campaign); err != nil {
		return err
	}
	campaign.Version++
	s.campaigns[id] = CloneCampaign(campaign)
	return nil
}

func (s *Store) RecordCallback(callback Callback) (bool, error) {
	key, err := CallbackKey(callback)
	if err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.callbacks[key]; exists {
		return false, nil
	}
	s.callbacks[key] = callback
	return true, nil
}

func (s *Store) SaveIdempotent(key string, body []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.idempotency[key]; exists {
		return false
	}
	s.idempotency[key] = append([]byte(nil), body...)
	return true
}

func (s *Store) Idempotent(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.idempotency[key]
	return append([]byte(nil), value...), ok
}

func (s *Store) PutSession(session Session) error {
	if session.Token == "" || session.TenantID == "" || session.UserID == "" {
		return fmt.Errorf("session identity is missing: %w", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.Token] = session
	return nil
}

func (s *Store) Session(token string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[token]
	return session, ok
}

func (s *Store) RevokeSession(token string, at time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[token]
	if !ok || session.RevokedAt != nil {
		return false
	}
	session.RevokedAt = &at
	s.sessions[token] = session
	return true
}

func (s *Store) AppendEvent(event Event) error {
	if event.TenantID == "" || event.RequestID == "" || event.ObjectID == "" || event.Action == "" {
		return fmt.Errorf("audit correlation is incomplete: %w", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *Store) QueryDevices(query Query) Page[Device] {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := make([]Device, 0)
	for _, device := range s.devices {
		if device.TenantID != query.TenantID || query.Class != "" && device.Class != query.Class {
			continue
		}
		filtered = append(filtered, device)
	}
	slices.SortFunc(filtered, func(left, right Device) int { return compareString(left.ID, right.ID) })
	total := len(filtered)
	start := min(max(query.Offset, 0), total)
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	end := min(start+limit, total)
	return Page[Device]{Items: CloneDevices(filtered[start:end]), Total: total}
}

func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
