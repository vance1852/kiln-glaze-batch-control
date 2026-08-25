package rollout

import "testing"

// TestSaveIdempotentPreservesFirstResponse reproduces the duplicate-submit bug:
// when a client retries with the same idempotency key but a different response
// body (e.g. a network retry landing on a different outcome), the store must
// keep the first (confirmed) response and replay it on subsequent reads. The
// buggy implementation overwrote the stored response, so later queries saw the
// second content instead of the originally confirmed wave.
func TestSaveIdempotentPreservesFirstResponse(t *testing.T) {
	store := NewStore()
	scope := "tenant-a\x00POST\x00/v1/rollout_waves\x00idem-1"

	first := []byte(`{"id":"wave-1"}`)
	if stored := store.SaveIdempotent(scope, first); !stored {
		t.Fatalf("first submit stored=%v, want true", stored)
	}

	// Retried request with the same key but different response content.
	second := []byte(`{"id":"wave-2"}`)
	if stored := store.SaveIdempotent(scope, second); stored {
		t.Fatalf("duplicate submit stored=%v, want false (replay, not new)", stored)
	}

	replayed, ok := store.Idempotent(scope)
	if !ok {
		t.Fatal("idempotent record missing after duplicate submit")
	}
	if string(replayed) != string(first) {
		t.Fatalf("replayed response=%s, want first %s", replayed, first)
	}
}
