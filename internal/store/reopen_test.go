package store

import (
	"path/filepath"
	"testing"
	"time"

	"activityregistration/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")
	now := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	event := domain.Event{ID: "event-reopen", Name: "Persistent Event", RegistrationOpensAt: now, Deadline: now.Add(time.Hour), Capacity: 4, Status: domain.EventOpen}
	registration := domain.Registration{ID: "reg-reopen", EventID: event.ID, Name: "Lin", Phone: "13800000001", SubmittedAt: now, Status: domain.RegistrationApproved}
	review := domain.ReviewRecord{ID: "review-reopen", EventID: event.ID, RegistrationID: registration.ID, Decision: domain.DecisionApprove, Reviewer: "reviewer", ReviewedAt: now}
	audit := domain.AuditEvent{ID: "audit-reopen", EventID: event.ID, RegistrationID: registration.ID, Action: "reopen-check", Actor: "test", OccurredAt: now}
	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, writeErr := range []error{first.PutEvent(event), first.PutRegistration(registration), first.PutReview(review), first.PutAuditEvent(audit)} {
		if writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if got, err := second.GetEvent(event.ID); err != nil || got.ID != event.ID {
		t.Fatalf("event did not survive reopen: %v", err)
	}
	if got, err := second.GetRegistration(event.ID, registration.ID); err != nil || got.Status != registration.Status {
		t.Fatalf("registration did not survive reopen: %v", err)
	}
	if got, err := second.GetReview(event.ID, review.ID); err != nil || got.Reviewer != review.Reviewer {
		t.Fatalf("review did not survive reopen: %v", err)
	}
	if events, err := second.ListAuditEvents(event.ID); err != nil || len(events) != 1 {
		t.Fatalf("audit did not survive reopen: %v", err)
	}
}
