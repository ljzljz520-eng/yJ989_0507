package service

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"activityregistration/internal/clock"
	"activityregistration/internal/domain"
	"activityregistration/internal/store"
)

func newTestService(t *testing.T) (*Service, *clock.FixedClock, func()) {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	current := clock.NewFixed(time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC))
	return New(database, current), current, func() { database.Close() }
}

func makeEvent(t *testing.T, application *Service) domain.Event {
	t.Helper()
	event, err := application.CreateEvent(EventInput{ID: "event-1", Name: "Community Day", RegistrationOpensAt: time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC), Deadline: time.Date(2025, 1, 10, 9, 0, 0, 0, time.UTC), Capacity: 100})
	if err != nil {
		t.Fatal(err)
	}
	event, err = application.PublishEvent(event.ID)
	if err != nil {
		t.Fatal(err)
	}
	return event
}

func TestServiceSubmitReviewExport(t *testing.T) {
	application, _, cleanup := newTestService(t)
	defer cleanup()
	event := makeEvent(t, application)
	registration, err := application.SubmitRegistration(event.ID, RegistrationInput{ID: "reg-1", Name: "Ada", Phone: "13800000000", Note: "access"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := application.SubmitRegistration(event.ID, RegistrationInput{ID: "reg-2", Name: "Other", Phone: "13800000000"}); !errors.Is(err, ErrDuplicatePhone) {
		t.Fatalf("expected duplicate phone, got %v", err)
	}
	review, err := application.ReviewRegistration(event.ID, registration.ID, ReviewInput{Decision: domain.DecisionApprove, Reviewer: "reviewer", Reason: "fits"})
	if err != nil || review.Decision != domain.DecisionApprove {
		t.Fatalf("review failed: %v", err)
	}
	result, err := application.ListRegistrations(event.ID, ListOptions{Page: 1, PageSize: 10})
	if err != nil || result.Total != 1 || result.Items[0].Status != domain.RegistrationApproved {
		t.Fatalf("list failed: %v", err)
	}
	content, job, err := application.ExportRegistrations(event.ID, "admin")
	if err != nil || job.RowCount != 1 || len(content) == 0 {
		t.Fatalf("export failed: %v", err)
	}
	if len(mustAudit(t, application, event.ID)) < 4 {
		t.Fatal("audit trail is incomplete")
	}
}

func mustAudit(t *testing.T, application *Service, eventID string) []domain.AuditEvent {
	t.Helper()
	items, err := application.AuditTrail(eventID)
	if err != nil {
		t.Fatal(err)
	}
	return items
}
