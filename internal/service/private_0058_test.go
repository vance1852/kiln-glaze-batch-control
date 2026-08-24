package service

import (
  "context"
  "errors"
  "testing"
  "firmware-rollout-control/internal/idempotency"
)

type privateIdempotencyStore0058 struct { journalErr error }
func (s *privateIdempotencyStore0058) Get(context.Context, string) (idempotency.Record, bool, error) { return idempotency.Record{}, false, nil }
func (s *privateIdempotencyStore0058) Put(context.Context, idempotency.Record) error { return s.journalErr }
func TestIdempotentCreateDoesNotHideJournalError(t *testing.T) {
  want := errors.New("idempotency journal unavailable")
  store := &privateIdempotencyStore0058{journalErr: want}
  code, body, err := New(nil).IdempotentCreate(context.Background(), store, "idem-58", []byte("body"), func() (int, any, error) { return 201, map[string]string{"id":"record-58"}, nil })
  if !errors.Is(err, want) { t.Fatalf("err=%v", err) }
  if code != 0 || body != nil { t.Fatalf("code=%d body=%v", code, body) }
}
