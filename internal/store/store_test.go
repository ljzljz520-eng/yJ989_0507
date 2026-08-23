package store

import (
	"path/filepath"
	"testing"
	"time"

	"activityregistration/internal/domain"
)

func TestStorePersistsBusinessRecords(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "activity.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	now := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	event := domain.Event{ID: "event-1", Name: "Open Day", RegistrationOpensAt: now, Deadline: now.Add(time.Hour), Capacity: 10, Status: domain.EventOpen}
	registration := domain.Registration{ID: "reg-1", EventID: event.ID, Name: "Ada", Phone: "13800000000", SubmittedAt: now, Status: domain.RegistrationPending}
	review := domain.ReviewRecord{ID: "review-1", EventID: event.ID, RegistrationID: registration.ID, Decision: domain.DecisionApprove, Reviewer: "admin", ReviewedAt: now}
	audit := domain.AuditEvent{ID: "audit-1", EventID: event.ID, RegistrationID: registration.ID, Action: "created", Actor: "admin", OccurredAt: now}
	if err := database.PutEvent(event); err != nil {
		t.Fatal(err)
	}
	if err := database.PutRegistration(registration); err != nil {
		t.Fatal(err)
	}
	if err := database.PutReview(review); err != nil {
		t.Fatal(err)
	}
	if err := database.PutAuditEvent(audit); err != nil {
		t.Fatal(err)
	}
	if got, err := database.GetEvent(event.ID); err != nil || got.Name != event.Name {
		t.Fatalf("event read failed: %v", err)
	}
	if got, err := database.GetRegistration(event.ID, registration.ID); err != nil || got.Phone != registration.Phone {
		t.Fatalf("registration read failed: %v", err)
	}
	if got, err := database.LatestReview(event.ID, registration.ID); err != nil || got.Decision != review.Decision {
		t.Fatalf("review read failed: %v", err)
	}
	if count, err := database.CountAuditEvents(event.ID); err != nil || count != 1 {
		t.Fatalf("audit count failed: %v", err)
	}
}
