package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"activityregistration/internal/domain"
)

type CollaborationSummary struct {
	EventID      string         `json:"event_id"`
	Actors       []string       `json:"actors"`
	ActionCounts map[string]int `json:"action_counts"`
	LatestAction string         `json:"latest_action"`
	LatestActor  string         `json:"latest_actor"`
	AuditCount   int            `json:"audit_count"`
}

func (service *Service) AssignReviewer(eventID, registrationID, reviewer, owner string) error {
	reviewer = domain.NormalizeActor(reviewer)
	owner = domain.NormalizeActor(owner)
	if reviewer == "" || owner == "" {
		return errors.New("reviewer and owner are required")
	}
	if reviewer == owner {
		return errors.New("owner cannot assign the same actor as a separate reviewer")
	}
	if _, err := service.store.GetRegistration(eventID, registrationID); err != nil {
		return err
	}
	return service.recordAudit(eventID, registrationID, "reviewer.assigned", owner, reviewer)
}

func (service *Service) AddCollaborationNote(note domain.CollaborationNote) error {
	validated, err := domain.NewCollaborationNote(note.EventID, note.RegistrationID, note.Actor, note.Text)
	if err != nil {
		return err
	}
	return service.recordAudit(validated.EventID, validated.RegistrationID, "collaboration.note", validated.Actor, validated.Text)
}

func (service *Service) CollaborationSummary(eventID string) (CollaborationSummary, error) {
	audits, err := service.store.ListAuditEvents(eventID)
	if err != nil {
		return CollaborationSummary{}, err
	}
	summary := CollaborationSummary{EventID: eventID, Actors: domain.UniqueActors(audits), ActionCounts: domain.ActionCounts(audits), AuditCount: len(audits)}
	if len(audits) == 0 {
		return summary, nil
	}
	sort.SliceStable(audits, func(left, right int) bool {
		if audits[left].OccurredAt.Equal(audits[right].OccurredAt) {
			return audits[left].ID < audits[right].ID
		}
		return audits[left].OccurredAt.Before(audits[right].OccurredAt)
	})
	latest := audits[len(audits)-1]
	summary.LatestAction = latest.Action
	summary.LatestActor = latest.Actor
	return summary, nil
}

func (service *Service) ActorCanReview(eventID, registrationID, actor string) (bool, error) {
	registration, err := service.store.GetRegistration(eventID, registrationID)
	if err != nil {
		return false, err
	}
	if registration.Status != domain.RegistrationPending {
		return false, nil
	}
	actor = domain.NormalizeActor(actor)
	if actor == "" {
		return false, nil
	}
	audits, err := service.store.ListAuditEvents(eventID)
	if err != nil {
		return false, err
	}
	for _, audit := range audits {
		if audit.RegistrationID == registrationID && audit.Action == "reviewer.assigned" && domain.NormalizeActor(audit.Detail) == actor {
			return true, nil
		}
	}
	return false, nil
}

func (service *Service) RecordOperatorNote(eventID, actor, text string) error {
	note, err := domain.NewCollaborationNote(eventID, "", actor, text)
	if err != nil {
		return err
	}
	return service.AddCollaborationNote(note)
}

func (service *Service) AuditByActor(eventID, actor string) ([]domain.AuditEvent, error) {
	audits, err := service.store.ListAuditEvents(eventID)
	if err != nil {
		return nil, err
	}
	needle := domain.NormalizeActor(actor)
	filtered := make([]domain.AuditEvent, 0)
	for _, audit := range audits {
		if domain.NormalizeActor(audit.Actor) == needle {
			filtered = append(filtered, audit)
		}
	}
	return filtered, nil
}

func (service *Service) ValidateWorkflowActors(eventID string, collaborators []domain.Collaborator) error {
	if len(collaborators) == 0 {
		return errors.New("at least one collaborator is required")
	}
	seen := make(map[string]bool)
	for _, collaborator := range collaborators {
		if err := domain.ValidateCollaborator(collaborator); err != nil {
			return err
		}
		actor := domain.NormalizeActor(collaborator.Actor)
		if seen[actor] {
			return fmt.Errorf("duplicate collaborator %s", actor)
		}
		seen[actor] = true
	}
	if _, err := service.store.GetEvent(eventID); err != nil {
		return err
	}
	return nil
}

func (service *Service) DescribeWorkflow(eventID string) (string, error) {
	event, err := service.store.GetEvent(eventID)
	if err != nil {
		return "", err
	}
	summary, err := service.EventSummary(eventID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s: %s, %d of %d registrations approved", event.Name, strings.ToLower(string(event.Status)), summary.Approved, summary.Total), nil
}
