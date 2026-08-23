package service

import (
	"strings"
	"testing"
)

func TestImportRegistrationReport(t *testing.T) {
	application, _, cleanup := newTestService(t)
	defer cleanup()
	event := makeEvent(t, application)
	report, err := application.ImportRegistrations(event.ID, strings.NewReader("name,phone,note\nAda,13800000001,first\nBob,bad,second\n"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Imported != 1 || len(report.Invalid) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
}
