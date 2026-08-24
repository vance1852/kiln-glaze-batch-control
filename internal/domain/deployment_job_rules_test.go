package domain

import (
	"errors"
	"testing"
	"time"
)

func TestActivationValidation(t *testing.T) {
	c := Activation{DeploymentJobID: "s1", To: "release_operator-2", Location: "lab", RecordedAt: time.Now().UTC()}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Activation{DeploymentJobID: "s1"}).Validate(); !errors.Is(err, ErrConflict) {
		t.Fatalf("missing activation fields = %v", err)
	}
}

func TestDeploymentJobInProgressRequiresAcceptedAndUnexpired(t *testing.T) {
	now := time.Now().UTC()
	s := DeploymentJob{Status: DeploymentJobAccepted, ExpiresAt: now.Add(time.Hour)}
	if err := s.CanBePerformed(now); err != nil {
		t.Fatal(err)
	}
	s.Status = DeploymentJobCompleted
	if err := s.CanBePerformed(now); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("status error = %v", err)
	}
	s.Status = DeploymentJobAccepted
	if err := s.CanBePerformed(now.Add(2 * time.Hour)); !errors.Is(err, ErrExpired) {
		t.Fatalf("expiry error = %v", err)
	}
}

func TestInstallationReportDecisionRejectsNegativeLimit(t *testing.T) {
	status, err := InstallationReportDecision(1, 2)
	if err != nil || status != InstallationReportVerified {
		t.Fatalf("decision = %s, %v", status, err)
	}
	if _, err := InstallationReportDecision(1, -1); !errors.Is(err, ErrConflict) {
		t.Fatalf("negative limit error = %v", err)
	}
}
