package domain

import (
	"testing"
	"time"
)

func TestValidateEventAndRegistration(t *testing.T) {
	opening := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	event := Event{ID: "event-1", Name: "Community Day", RegistrationOpensAt: opening, Deadline: opening.Add(24 * time.Hour), Capacity: 20, Status: EventDraft}
	if err := ValidateEvent(event); err != nil {
		t.Fatal(err)
	}
	registration := Registration{ID: "reg-1", EventID: event.ID, Name: "Ada", Phone: "+86 138-0000-0000", Note: "access needs"}
	if err := ValidateRegistration(registration); err != nil {
		t.Fatal(err)
	}
	if NormalizePhone(registration.Phone) != "+8613800000000" {
		t.Fatalf("unexpected normalized phone")
	}
	if err := CanAcceptRegistration(event, 0, true); err == nil {
		t.Fatal("draft event should not accept registrations")
	}
}

func TestEventTransitions(t *testing.T) {
	event := Event{ID: "event-1", Name: "Day", RegistrationOpensAt: time.Unix(0, 0), Deadline: time.Unix(100, 0), Capacity: 1, Status: EventDraft}
	if err := TransitionEvent(event, EventOpen); err != nil {
		t.Fatal(err)
	}
	if err := TransitionEvent(event, EventClosed); err == nil {
		t.Fatal("draft cannot close directly")
	}
	if err := ValidateReview(DecisionApprove, "reviewer"); err != nil {
		t.Fatal(err)
	}
}
