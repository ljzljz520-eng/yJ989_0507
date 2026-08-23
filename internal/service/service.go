package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"activityregistration/internal/clock"
	"activityregistration/internal/domain"
	"activityregistration/internal/store"
)

var (
	ErrDuplicatePhone = errors.New("phone already registered")
	ErrBatchFinished  = errors.New("batch is already finished")
	ErrBatchStep      = errors.New("batch is at an unexpected step")
)

type EventInput struct {
	ID                  string
	Name                string
	RegistrationOpensAt time.Time
	Deadline            time.Time
	Capacity            int
}

type RegistrationInput struct {
	ID    string
	Name  string
	Phone string
	Note  string
}

type ReviewInput struct {
	Decision domain.ReviewDecision
	Reason   string
	Reviewer string
}

type ListOptions struct {
	Page     int
	PageSize int
	Status   domain.RegistrationStatus
}

type ListResult struct {
	Items      []domain.Registration `json:"items"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	Total      int                   `json:"total"`
	TotalPages int                   `json:"total_pages"`
}

type ImportReport struct {
	Imported int
	Invalid  []string
}

type Sequence struct {
	prefix string
	next   int
}

func NewSequence(prefix string) *Sequence {
	return &Sequence{prefix: prefix, next: 1}
}

func (sequence *Sequence) Next() string {
	value := fmt.Sprintf("%s-%04d", sequence.prefix, sequence.next)
	sequence.next++
	return value
}

type Service struct {
	store     *store.Store
	clock     clock.Clock
	eventIDs  *Sequence
	regIDs    *Sequence
	reviewIDs *Sequence
	auditIDs  *Sequence
	batchIDs  *Sequence
	exportIDs *Sequence
}

func New(database *store.Store, current clock.Clock) *Service {
	return &Service{
		store:     database,
		clock:     current,
		eventIDs:  NewSequence("event"),
		regIDs:    NewSequence("reg"),
		reviewIDs: NewSequence("review"),
		auditIDs:  NewSequence("audit"),
		batchIDs:  NewSequence("batch"),
		exportIDs: NewSequence("export"),
	}
}

func (service *Service) CreateEvent(input EventInput) (domain.Event, error) {
	if input.ID == "" {
		input.ID = service.eventIDs.Next()
	}
	event := domain.Event{
		ID:                  input.ID,
		Name:                strings.TrimSpace(input.Name),
		RegistrationOpensAt: input.RegistrationOpensAt.UTC(),
		Deadline:            input.Deadline.UTC(),
		Capacity:            input.Capacity,
		Status:              domain.EventDraft,
		CreatedAt:           service.clock.Now(),
		UpdatedAt:           service.clock.Now(),
	}
	if err := domain.ValidateEvent(event); err != nil {
		return domain.Event{}, err
	}
	if _, err := service.store.GetEvent(event.ID); err == nil {
		return domain.Event{}, fmt.Errorf("event %s already exists", event.ID)
	}
	if err := service.store.PutEvent(event); err != nil {
		return domain.Event{}, err
	}
	if err := service.recordAudit(event.ID, "", "event.created", "system", event.Name); err != nil {
		return domain.Event{}, err
	}
	return event, nil
}

func (service *Service) GetEvent(id string) (domain.Event, error) {
	return service.store.GetEvent(id)
}

func (service *Service) ListEvents() ([]domain.Event, error) {
	events, err := service.store.ListEvents()
	if err != nil {
		return nil, err
	}
	for index := range events {
		if events[index].Status == "" {
			events[index].Status = domain.EventDraft
		}
	}
	return events, nil
}

func (service *Service) SearchEvents(query string) ([]domain.Event, error) {
	events, err := service.ListEvents()
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return events, nil
	}
	filtered := make([]domain.Event, 0)
	for _, event := range events {
		if strings.Contains(strings.ToLower(event.Name), needle) || strings.Contains(strings.ToLower(event.ID), needle) {
			filtered = append(filtered, event)
		}
	}
	return filtered, nil
}

func (service *Service) UpdateEvent(input EventInput) (domain.Event, error) {
	event, err := service.store.GetEvent(input.ID)
	if err != nil {
		return domain.Event{}, err
	}
	if event.Status == domain.EventArchived {
		return domain.Event{}, domain.ErrInvalidTransition
	}
	if strings.TrimSpace(input.Name) != "" {
		event.Name = strings.TrimSpace(input.Name)
	}
	if !input.RegistrationOpensAt.IsZero() {
		event.RegistrationOpensAt = input.RegistrationOpensAt.UTC()
	}
	if !input.Deadline.IsZero() {
		event.Deadline = input.Deadline.UTC()
	}
	if input.Capacity > 0 {
		event.Capacity = input.Capacity
	}
	event.UpdatedAt = service.clock.Now()
	if err := domain.ValidateEvent(event); err != nil {
		return domain.Event{}, err
	}
	if err := service.store.PutEvent(event); err != nil {
		return domain.Event{}, err
	}
	return event, service.recordAudit(event.ID, "", "event.updated", "admin", event.Name)
}

func (service *Service) ChangeEventStatus(eventID string, target domain.EventStatus, actor string) (domain.Event, error) {
	event, err := service.store.GetEvent(eventID)
	if err != nil {
		return domain.Event{}, err
	}
	if err := domain.TransitionEvent(event, target); err != nil {
		return domain.Event{}, err
	}
	event.Status = target
	event.UpdatedAt = service.clock.Now()
	if err := service.store.PutEvent(event); err != nil {
		return domain.Event{}, err
	}
	return event, service.recordAudit(event.ID, "", "event.status."+string(target), actor, string(target))
}

func (service *Service) PublishEvent(eventID string) (domain.Event, error) {
	return service.ChangeEventStatus(eventID, domain.EventOpen, "admin")
}

func (service *Service) CloseEvent(eventID string) (domain.Event, error) {
	return service.ChangeEventStatus(eventID, domain.EventClosed, "admin")
}

func (service *Service) ArchiveEvent(eventID string) (domain.Event, error) {
	return service.ChangeEventStatus(eventID, domain.EventArchived, "admin")
}

func (service *Service) SubmitRegistration(eventID string, input RegistrationInput) (domain.Registration, error) {
	event, err := service.store.GetEvent(eventID)
	if err != nil {
		return domain.Registration{}, err
	}
	count, err := service.store.CountRegistrations(eventID)
	if err != nil {
		return domain.Registration{}, err
	}
	if err := domain.CanAcceptRegistration(event, count, service.clock.Now().Before(event.Deadline)); err != nil {
		return domain.Registration{}, err
	}
	phone := domain.NormalizePhone(input.Phone)
	if _, findErr := service.store.FindRegistrationByPhone(eventID, phone); findErr == nil {
		return domain.Registration{}, ErrDuplicatePhone
	} else if !errors.Is(findErr, store.ErrNotFound) {
		return domain.Registration{}, findErr
	}
	if input.ID == "" {
		input.ID = service.regIDs.Next()
	}
	registration := domain.Registration{
		ID:          input.ID,
		EventID:     eventID,
		Name:        strings.TrimSpace(input.Name),
		Phone:       phone,
		Note:        strings.TrimSpace(input.Note),
		SubmittedAt: service.clock.Now(),
		Status:      domain.RegistrationPending,
		UpdatedAt:   service.clock.Now(),
	}
	registration.Score = domain.ScoreRegistration(registration)
	if err := domain.ValidateRegistration(registration); err != nil {
		return domain.Registration{}, err
	}
	if err := service.store.PutRegistration(registration); err != nil {
		return domain.Registration{}, err
	}
	if err := service.recordAudit(eventID, registration.ID, "registration.submitted", "user", registration.Name); err != nil {
		return domain.Registration{}, err
	}
	return registration, nil
}

func (service *Service) GetRegistration(eventID, registrationID string) (domain.Registration, error) {
	return service.store.GetRegistration(eventID, registrationID)
}

func (service *Service) ReviewRegistration(eventID, registrationID string, input ReviewInput) (domain.ReviewRecord, error) {
	if err := domain.ValidateReview(input.Decision, input.Reviewer); err != nil {
		return domain.ReviewRecord{}, err
	}
	registration, err := service.GetRegistration(eventID, registrationID)
	if err != nil {
		return domain.ReviewRecord{}, err
	}
	if registration.Status != domain.RegistrationPending {
		return domain.ReviewRecord{}, fmt.Errorf("registration %s is not pending", registrationID)
	}
	status := domain.RegistrationRejected
	if input.Decision == domain.DecisionApprove {
		status = domain.RegistrationApproved
	}
	if _, err := service.store.UpdateRegistrationStatus(eventID, registrationID, status, input.Reviewer); err != nil {
		return domain.ReviewRecord{}, err
	}
	review := domain.ReviewRecord{ID: service.reviewIDs.Next(), EventID: eventID, RegistrationID: registrationID, Decision: input.Decision, Reason: strings.TrimSpace(input.Reason), Reviewer: strings.TrimSpace(input.Reviewer), ReviewedAt: service.clock.Now()}
	if err := service.store.PutReview(review); err != nil {
		return domain.ReviewRecord{}, err
	}
	if err := service.recordAudit(eventID, registrationID, "registration.reviewed", input.Reviewer, string(input.Decision)); err != nil {
		return domain.ReviewRecord{}, err
	}
	return review, nil
}

func (service *Service) ListReviews(eventID, registrationID string) ([]domain.ReviewRecord, error) {
	return service.store.ListReviews(eventID, registrationID)
}

func (service *Service) recordAudit(eventID, registrationID, action, actor, detail string) error {
	return service.store.PutAuditEvent(domain.AuditEvent{ID: service.auditIDs.Next(), EventID: eventID, RegistrationID: registrationID, Action: action, Actor: actor, Detail: detail, OccurredAt: service.clock.Now()})
}

func (service *Service) AuditTrail(eventID string) ([]domain.AuditEvent, error) {
	return service.store.ListAuditEvents(eventID)
}
