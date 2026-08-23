package service_test

import (
	"testing"

	"firmware-rollout-control/internal/rollout"
	"firmware-rollout-control/internal/service"
)

func TestIdempotentResponseKeepsFirstBody(t *testing.T) {
	control := service.NewFirmwareControl(rollout.NewStore())
	if saved, err := control.SaveIdempotentResponse("tenant-18", "POST", "/waves", "request-18", []byte(`{"wave":"first"}`)); err != nil || !saved {
		t.Fatalf("first save saved=%v err=%v", saved, err)
	}
	if saved, err := control.SaveIdempotentResponse("tenant-18", "POST", "/waves", "request-18", []byte(`{"wave":"duplicate"}`)); err != nil || saved {
		t.Fatalf("duplicate save saved=%v err=%v", saved, err)
	}
	body, ok, err := control.LoadIdempotentResponse("tenant-18", "POST", "/waves", "request-18")
	if err != nil || !ok || string(body) != `{"wave":"first"}` {
		t.Fatalf("stored response=%s ok=%v err=%v", body, ok, err)
	}
}
