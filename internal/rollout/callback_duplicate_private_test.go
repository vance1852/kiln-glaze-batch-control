package rollout

import "testing"

func TestDuplicateCallbackDoesNotReportCreated(t *testing.T) {
 s := NewStore(); c := Callback{TenantID:"t",DeviceID:"d",ArtifactID:"a",EventID:"e"}
 first, err := s.RecordCallback(c); if err != nil || !first { t.Fatalf("first=%v err=%v", first, err) }
 second, err := s.RecordCallback(c); if err != nil || second { t.Fatalf("duplicate=%v err=%v", second, err) }
}
