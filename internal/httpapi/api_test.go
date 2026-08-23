package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"activityregistration/internal/clock"
	"activityregistration/internal/service"
	"activityregistration/internal/store"
)

func TestHTTPEventAndRegistrationRoutes(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "http.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	application := service.New(database, clock.NewFixed(time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)))
	handler := NewHandler(application)
	create := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{"id":"event-http","name":"HTTP Day","registration_opens_at":"2024-12-01T09:00:00Z","deadline":"2025-01-10T09:00:00Z","capacity":5}`))
	handler.ServeHTTP(create, request)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status: %d body=%s", create.Code, create.Body.String())
	}
	publish := httptest.NewRecorder()
	handler.ServeHTTP(publish, httptest.NewRequest(http.MethodPost, "/events/event-http/publish", nil))
	if publish.Code != http.StatusOK {
		t.Fatalf("publish status: %d", publish.Code)
	}
	registration := httptest.NewRecorder()
	handler.ServeHTTP(registration, httptest.NewRequest(http.MethodPost, "/events/event-http/registrations", strings.NewReader(`{"name":"Ada","phone":"13800000000","note":"hello"}`)))
	if registration.Code != http.StatusCreated {
		t.Fatalf("registration status: %d body=%s", registration.Code, registration.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(registration.Body.Bytes(), &payload); err != nil || payload["event_id"] != "event-http" {
		t.Fatalf("unexpected registration response: %s", registration.Body.String())
	}
}
