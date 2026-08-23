package service

import (
	"context"
	"fmt"

	auditpkg "firmware-rollout-control/internal/audit"
	"firmware-rollout-control/internal/repository"
)

func ValidateAuditEvent(event auditpkg.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("audit event: %w", err)
	}
	_, err := event.JSON()
	return err
}

func (s *Service) WriteAuditEvent(ctx context.Context, event auditpkg.Event) error {
	if event.Detail == nil {
		event.Detail = make(map[string]any)
	}
	_, err := s.auditChain.Advance(event, func(head string) error {
		detail := event.Detail
		detail["chain_head"] = head
		payload, err := event.JSON()
		if err != nil {
			return err
		}
		return s.repo.WriteAudit(ctx, auditInput(event, payload))
	})
	if err != nil {
		return fmt.Errorf("advance audit chain: %w", err)
	}
	return nil
}

func auditInput(event auditpkg.Event, detail []byte) repository.AuditInput {
	return repository.AuditInput{RequestID: event.RequestID, ReleaseOperatorID: event.ReleaseOperatorID, ObjectType: event.ObjectType, ObjectID: event.ObjectID, Action: event.Action, Outcome: event.Outcome, Detail: detail}
}
