package domain

import "testing"

func TestNormalizeLabelsCopiesAndCleansValues(t *testing.T) {
	labels := map[string]string{" Managed_Device ": " North ", "": "ignored", "empty": " "}
	got := NormalizeLabels(labels)
	if got["managed_device"] != "North" || len(got) != 1 {
		t.Fatalf("labels=%v", got)
	}
	got["managed_device"] = "changed"
	if labels[" Managed_Device "] != " North " {
		t.Fatal("source labels were mutated")
	}
}
