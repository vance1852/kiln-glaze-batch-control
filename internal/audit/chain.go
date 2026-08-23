package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

type Chain struct {
	mu       sync.Mutex
	previous string
}

// Append advances the chain by one event and returns the new head. It is a
// convenience for purely in-memory chains. Production callers that persist the
// event should use Advance so the head only commits when the write succeeds.
func (c *Chain) Append(event Event) (string, error) {
	return c.advance(event, nil)
}

// Advance computes the next chain head for event and invokes persist with it.
// The chain head advances only when persist returns nil, so a failed write
// leaves the chain unchanged: every persisted event receives a unique head
// linked into the strict chain order, and no head is ever recorded without a
// matching row. The chain is locked for the whole call, serializing concurrent
// advances and eliminating the read-modify-write race on the shared head.
func (c *Chain) Advance(event Event, persist func(head string) error) (string, error) {
	return c.advance(event, persist)
}

func (c *Chain) advance(event Event, persist func(head string) error) (string, error) {
	if err := event.Validate(); err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	head, err := c.computeLocked(event)
	if err != nil {
		return "", err
	}
	if persist != nil {
		if err := persist(head); err != nil {
			return "", err
		}
	}
	c.previous = head
	return head, nil
}

func (c *Chain) computeLocked(event Event) (string, error) {
	if event.Detail == nil {
		event.Detail = map[string]any{}
	}
	body, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("marshal audit event: %w", err)
	}
	payload, err := json.Marshal(struct {
		Previous string          `json:"previous"`
		Event    json.RawMessage `json:"event"`
	}{Previous: c.previous, Event: body})
	if err != nil {
		return "", fmt.Errorf("marshal audit chain: %w", err)
	}
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:]), nil
}

func (c *Chain) Head() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.previous
}
