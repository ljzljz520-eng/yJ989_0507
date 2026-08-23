package exporter

import (
	"strings"
	"testing"
	"time"

	"activityregistration/internal/domain"
)

func TestRenderRegistrations(t *testing.T) {
	content, err := RenderRegistrations([]domain.Registration{{ID: "reg-1", EventID: "event-1", Name: "Ada, A", Phone: "13800000000", Note: "bring badge", Score: 60, SubmittedAt: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC), Status: domain.RegistrationApproved}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "id,event_id,name") || !strings.Contains(content, "\"Ada, A\"") {
		t.Fatalf("unexpected csv: %s", content)
	}
	if Summary([]domain.Registration{{Status: domain.RegistrationApproved}, {Status: domain.RegistrationPending}})["total"] != 2 {
		t.Fatal("summary total mismatch")
	}
}
