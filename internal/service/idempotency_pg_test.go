package service

import (
	"context"
	"errors"
	"testing"

	"firmware-rollout-control/internal/idempotency"
)

type failingPutStore struct {
	getErr  error
	putErr  error
	record  idempotency.Record
	found   bool
	putCall int
}

func (s *failingPutStore) Get(context.Context, string) (idempotency.Record, bool, error) {
	return s.record, s.found, s.getErr
}

func (s *failingPutStore) Put(_ context.Context, record idempotency.Record) error {
	s.putCall++
	if s.putErr != nil {
		return s.putErr
	}
	s.record, s.found = record, true
	return nil
}

func TestIdempotentCreateSurfacesPersistenceFailure(t *testing.T) {
	svc := New(nil)
	store := &failingPutStore{putErr: errors.New("disk full")}
	create := func() (int, any, error) { return 201, map[string]string{"id": "task-1"}, nil }
	code, payload, err := svc.IdempotentCreate(t.Context(), store, "key-9", []byte(`{"code":"S-9"}`), create)
	if err == nil {
		t.Fatal("idempotency persistence failure was swallowed")
	}
	if !errors.Is(err, store.putErr) {
		t.Fatalf("error does not wrap underlying cause: %v", err)
	}
	if code != 0 || payload != nil {
		t.Fatalf("unexpected success path code=%d payload=%s", code, payload)
	}
	if store.putCall != 1 {
		t.Fatalf("put calls=%d", store.putCall)
	}
}

func TestIdempotentCreateReplaysAfterSuccessfulPersist(t *testing.T) {
	svc := New(nil)
	store := &failingPutStore{}
	create := func() (int, any, error) { return 201, map[string]string{"id": "task-2"}, nil }
	if _, _, err := svc.IdempotentCreate(t.Context(), store, "key-10", []byte(`{"code":"S-10"}`), create); err != nil {
		t.Fatal(err)
	}
	code, payload, err := svc.IdempotentCreate(t.Context(), store, "key-10", []byte(`{"code":"S-10"}`), create)
	if err != nil {
		t.Fatalf("replay error=%v", err)
	}
	if code != 201 || string(payload) == "" {
		t.Fatalf("replay code=%d payload=%s", code, payload)
	}
	if store.putCall != 1 {
		t.Fatalf("create invoked after replay: put calls=%d", store.putCall)
	}
}
