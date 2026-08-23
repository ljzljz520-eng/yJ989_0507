package integration

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"activityregistration/internal/clock"
	"activityregistration/internal/domain"
	"activityregistration/internal/service"
	"activityregistration/internal/store"
)

func integrationService(t *testing.T) *service.Service {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "integration.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return service.New(database, clock.NewFixed(time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)))
}

func integrationEvent(t *testing.T, application *service.Service) domain.Event {
	t.Helper()
	event, err := application.CreateEvent(service.EventInput{ID: "workflow-event", Name: "Workflow Day", RegistrationOpensAt: time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC), Deadline: time.Date(2025, 2, 1, 9, 0, 0, 0, time.UTC), Capacity: 20})
	if err != nil {
		t.Fatal(err)
	}
	event, err = application.PublishEvent(event.ID)
	if err != nil {
		t.Fatal(err)
	}
	return event
}

func TestWorkflowCreateReviewArchive(t *testing.T) {
	application := integrationService(t)
	event := integrationEvent(t, application)
	registration, err := application.SubmitRegistration(event.ID, service.RegistrationInput{ID: "workflow-reg", Name: "Ada", Phone: "13800000000"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := application.ReviewRegistration(event.ID, registration.ID, service.ReviewInput{Decision: domain.DecisionApprove, Reviewer: "reviewer"}); err != nil {
		t.Fatal(err)
	}
	if _, err := application.CloseEvent(event.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := application.ArchiveEvent(event.ID); err != nil {
		t.Fatal(err)
	}
	archived, err := application.GetEvent(event.ID)
	if err != nil || archived.Status != domain.EventArchived {
		t.Fatalf("archive failed: %v", err)
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	application := integrationService(t)
	event, err := application.CreateEvent(service.EventInput{ID: "search-event", Name: "Original Day", RegistrationOpensAt: time.Date(2024, 12, 1, 9, 0, 0, 0, time.UTC), Deadline: time.Date(2025, 2, 1, 9, 0, 0, 0, time.UTC), Capacity: 20})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := application.UpdateEvent(service.EventInput{ID: event.ID, Name: "Published Day", Capacity: 30}); err != nil {
		t.Fatal(err)
	}
	items, err := application.SearchEvents("published")
	if err != nil || len(items) != 1 {
		t.Fatalf("search failed: %v", err)
	}
	if _, err := application.PublishEvent(event.ID); err != nil {
		t.Fatal(err)
	}
	got, err := application.GetEvent(event.ID)
	if err != nil || got.Status != domain.EventOpen {
		t.Fatalf("publish failed: %v", err)
	}
}

func TestWorkflowImportReport(t *testing.T) {
	application := integrationService(t)
	event := integrationEvent(t, application)
	report, err := application.ImportRegistrations(event.ID, strings.NewReader("name,phone,note\nAda,13800000000,first\nBob,13800000001,second\n"))
	if err != nil || report.Imported != 2 || len(report.Invalid) != 0 {
		t.Fatalf("import failed: %+v %v", report, err)
	}
	result, err := application.ListRegistrations(event.ID, service.ListOptions{Page: 1, PageSize: 10})
	if err != nil || result.Total != 2 {
		t.Fatalf("report total failed: %v", err)
	}
	if _, _, err := application.ExportRegistrations(event.ID, "reporter"); err != nil {
		t.Fatal(err)
	}
}
