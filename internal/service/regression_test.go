package service

import (
	"testing"
	"time"

	"activityregistration/internal/domain"
)

func TestBusiness014Regression(t *testing.T) {
	application, current, cleanup := newTestService(t)
	defer cleanup()
	event := makeEvent(t, application)
	for number := 60; number >= 1; number-- {
		_, err := application.SubmitRegistration(event.ID, RegistrationInput{ID: formatRegistrationID(number), Name: "A", Phone: formatPhone(number), Note: ""})
		if err != nil {
			t.Fatal(err)
		}
		current.Advance(time.Minute)
	}
	if _, err := application.CreateBatch(event.ID, "N-114"); err != nil {
		t.Fatal(err)
	}
	if _, err := application.AdvanceBatch(event.ID, "N-114"); err != nil {
		t.Fatal(err)
	}
	if _, err := application.AdvanceBatch(event.ID, "N-114"); err != nil {
		t.Fatal(err)
	}
	if _, err := application.RequestBatchCancellation(event.ID, "N-114", "operator"); err != nil {
		t.Fatal(err)
	}
	batch, err := application.ConfirmBatch(event.ID, "N-114", "operator")
	if err != nil || batch.Step != domain.BatchCancelled {
		t.Fatalf("batch cancellation was not confirmed: %v", err)
	}
	page, err := application.ListRegistrations(event.ID, ListOptions{Page: 2, PageSize: 30})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 60 {
		t.Fatalf("expected total to remain 60, got %d", page.Total)
	}
	for offset, registration := range page.Items {
		expected := formatRegistrationID(30 - offset)
		if registration.ID != expected {
			t.Fatalf("page two lost continuity at offset %d: got %s want %s", offset, registration.ID, expected)
		}
	}
}

func formatRegistrationID(number int) string {
	return "r-" + pad(number)
}

func formatPhone(number int) string {
	return "1380000" + pad(number)
}

func pad(number int) string {
	if number < 10 {
		return "00" + string(rune('0'+number))
	}
	if number < 100 {
		return "0" + string(rune('0'+number/10)) + string(rune('0'+number%10))
	}
	return string(rune('0'+number/100)) + string(rune('0'+(number/10)%10)) + string(rune('0'+number%10))
}
