package rollout

import "testing"

func TestUnauthorizedErrorClassification(t *testing.T) {
	status, code := ClassifyError(ErrUnauthorized)
	if status != 403 || code != "forbidden" {
		t.Fatalf("status=%d code=%s, authorization failures were misclassified", status, code)
	}
}
