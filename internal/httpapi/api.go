package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"activityregistration/internal/domain"
	"activityregistration/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(application *service.Service) *Handler {
	return &Handler{service: application}
}

func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	segments := pathSegments(request.URL.Path)
	if len(segments) == 0 {
		writeJSON(response, http.StatusOK, map[string]string{"service": "activity-registration", "status": "ready"})
		return
	}
	if segments[0] != "events" {
		writeError(response, http.StatusNotFound, errors.New("route not found"))
		return
	}
	if len(segments) == 1 {
		handler.events(response, request)
		return
	}
	handler.eventRoute(response, request, segments[1], segments[2:])
}

func pathSegments(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] == "" {
		return nil
	}
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func (handler *Handler) events(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		items, err := handler.service.SearchEvents(request.URL.Query().Get("q"))
		if err != nil {
			writeError(response, http.StatusInternalServerError, err)
			return
		}
		writeJSON(response, http.StatusOK, map[string]any{"items": items, "total": len(items)})
	case http.MethodPost:
		var input eventPayload
		if !decodeJSON(response, request, &input) {
			return
		}
		created, err := handler.service.CreateEvent(input.toService())
		if err != nil {
			writeError(response, http.StatusBadRequest, err)
			return
		}
		writeJSON(response, http.StatusCreated, created)
	default:
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (handler *Handler) eventRoute(response http.ResponseWriter, request *http.Request, eventID string, rest []string) {
	if len(rest) == 0 {
		handler.event(response, request, eventID)
		return
	}
	switch rest[0] {
	case "publish", "close", "archive":
		handler.statusChange(response, request, eventID, rest[0])
	case "summary":
		handler.summary(response, request, eventID)
	case "reviewers":
		handler.reviewers(response, request, eventID)
	case "search":
		handler.search(response, request, eventID)
	case "jobs":
		handler.jobs(response, request, eventID)
	case "collaboration":
		handler.collaboration(response, request, eventID)
	case "description":
		handler.description(response, request, eventID)
	case "registrations":
		handler.registrations(response, request, eventID, rest[1:])
	case "export":
		handler.export(response, request, eventID)
	case "batch":
		handler.batch(response, request, eventID, rest[1:])
	default:
		writeError(response, http.StatusNotFound, errors.New("route not found"))
	}
}

func (handler *Handler) event(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method == http.MethodDelete {
		if err := handler.service.DeleteEvent(eventID); err != nil {
			writeError(response, http.StatusBadRequest, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
		return
	}
	if request.Method == http.MethodGet {
		event, err := handler.service.GetEvent(eventID)
		if err != nil {
			writeStoreError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, event)
		return
	}
	if request.Method != http.MethodPut {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var input eventPayload
	if !decodeJSON(response, request, &input) {
		return
	}
	input.ID = eventID
	updated, err := handler.service.UpdateEvent(input.toService())
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}
	writeJSON(response, http.StatusOK, updated)
}

func (handler *Handler) statusChange(response http.ResponseWriter, request *http.Request, eventID, action string) {
	if request.Method != http.MethodPost {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	status := map[string]domain.EventStatus{"publish": domain.EventOpen, "close": domain.EventClosed, "archive": domain.EventArchived}[action]
	event, err := handler.service.ChangeEventStatus(eventID, status, request.Header.Get("X-Actor"))
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}
	writeJSON(response, http.StatusOK, event)
}

func (handler *Handler) registrations(response http.ResponseWriter, request *http.Request, eventID string, rest []string) {
	if len(rest) == 2 && (rest[1] == "cancel" || rest[1] == "restore") {
		handler.registrationAction(response, request, eventID, rest[0], rest[1])
		return
	}
	if len(rest) > 0 && rest[0] == "review" {
		if len(rest) != 2 {
			writeError(response, http.StatusNotFound, errors.New("registration id is required"))
			return
		}
		handler.review(response, request, eventID, rest[1])
		return
	}
	switch request.Method {
	case http.MethodGet:
		page, pageSize := pageParams(request)
		result, err := handler.service.ListRegistrations(eventID, service.ListOptions{Page: page, PageSize: pageSize, Status: domain.RegistrationStatus(request.URL.Query().Get("status"))})
		if err != nil {
			writeError(response, http.StatusInternalServerError, err)
			return
		}
		writeJSON(response, http.StatusOK, result)
	case http.MethodPost:
		var input registrationPayload
		if !decodeJSON(response, request, &input) {
			return
		}
		created, err := handler.service.SubmitRegistration(eventID, service.RegistrationInput{ID: input.ID, Name: input.Name, Phone: input.Phone, Note: input.Note})
		if err != nil {
			writeError(response, http.StatusBadRequest, err)
			return
		}
		writeJSON(response, http.StatusCreated, created)
	default:
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (handler *Handler) review(response http.ResponseWriter, request *http.Request, eventID, registrationID string) {
	if request.Method != http.MethodPost {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var input reviewPayload
	if !decodeJSON(response, request, &input) {
		return
	}
	review, err := handler.service.ReviewRegistration(eventID, registrationID, service.ReviewInput{Decision: domain.ReviewDecision(input.Decision), Reason: input.Reason, Reviewer: input.Reviewer})
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}
	writeJSON(response, http.StatusOK, review)
}

func (handler *Handler) summary(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method != http.MethodGet {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	summary, err := handler.service.EventSummary(eventID)
	if err != nil {
		writeStoreError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, summary)
}

func (handler *Handler) reviewers(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method != http.MethodGet {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	queue, err := handler.service.ReviewerQueue(eventID, limit)
	if err != nil {
		writeStoreError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, queue)
}

func (handler *Handler) search(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method != http.MethodGet {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	page, pageSize := pageParams(request)
	result, err := handler.service.SearchRegistrations(eventID, request.URL.Query().Get("q"), page, pageSize)
	if err != nil {
		writeStoreError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, result)
}

func (handler *Handler) jobs(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method != http.MethodGet {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	jobs, err := handler.service.ExportJobs(eventID)
	if err != nil {
		writeStoreError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{"items": jobs, "total": len(jobs)})
}

func (handler *Handler) collaboration(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method != http.MethodGet {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	summary, err := handler.service.CollaborationSummary(eventID)
	if err != nil {
		writeStoreError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, summary)
}

func (handler *Handler) description(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method != http.MethodGet {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	description, err := handler.service.DescribeWorkflow(eventID)
	if err != nil {
		writeStoreError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, map[string]string{"description": description})
}

func (handler *Handler) registrationAction(response http.ResponseWriter, request *http.Request, eventID, registrationID, action string) {
	if request.Method != http.MethodPost {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	actor := request.Header.Get("X-Actor")
	var registration domain.Registration
	var err error
	if action == "cancel" {
		registration, err = handler.service.CancelRegistration(eventID, registrationID, actor)
	} else {
		registration, err = handler.service.RestoreRegistration(eventID, registrationID, actor)
	}
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}
	writeJSON(response, http.StatusOK, registration)
}

func (handler *Handler) export(response http.ResponseWriter, request *http.Request, eventID string) {
	if request.Method != http.MethodGet {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	content, job, err := handler.service.ExportRegistrations(eventID, request.Header.Get("X-Actor"))
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}
	response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	response.Header().Set("X-Export-Job", job.ID)
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write([]byte(content))
}

func (handler *Handler) batch(response http.ResponseWriter, request *http.Request, eventID string, rest []string) {
	if request.Method != http.MethodPost {
		writeError(response, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	if len(rest) == 0 {
		var input struct {
			ID string `json:"id"`
		}
		if !decodeJSON(response, request, &input) {
			return
		}
		batch, err := handler.service.CreateBatch(eventID, input.ID)
		if err != nil {
			writeError(response, http.StatusBadRequest, err)
			return
		}
		writeJSON(response, http.StatusCreated, batch)
		return
	}
	batchID := rest[0]
	if len(rest) == 1 {
		batch, err := handler.service.AdvanceBatch(eventID, batchID)
		if err != nil {
			writeError(response, http.StatusBadRequest, err)
			return
		}
		writeJSON(response, http.StatusOK, batch)
		return
	}
	if rest[1] == "cancel" {
		batch, err := handler.service.RequestBatchCancellation(eventID, batchID, request.Header.Get("X-Actor"))
		if err != nil {
			writeError(response, http.StatusBadRequest, err)
			return
		}
		writeJSON(response, http.StatusOK, batch)
		return
	}
	if rest[1] == "confirm" {
		batch, err := handler.service.ConfirmBatch(eventID, batchID, request.Header.Get("X-Actor"))
		if err != nil {
			writeError(response, http.StatusBadRequest, err)
			return
		}
		writeJSON(response, http.StatusOK, batch)
		return
	}
	writeError(response, http.StatusNotFound, errors.New("batch action not found"))
}

func pageParams(request *http.Request) (int, int) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(request.URL.Query().Get("page_size"))
	return page, pageSize
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		writeError(response, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeError(response http.ResponseWriter, status int, err error) {
	writeJSON(response, status, map[string]string{"error": err.Error()})
}

func writeStoreError(response http.ResponseWriter, err error) {
	if service.IsNotFound(err) {
		writeError(response, http.StatusNotFound, err)
		return
	}
	writeError(response, http.StatusInternalServerError, err)
}
