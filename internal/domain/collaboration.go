package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type CollaboratorRole string

const (
	RoleOwner    CollaboratorRole = "owner"
	RoleReviewer CollaboratorRole = "reviewer"
	RoleOperator CollaboratorRole = "operator"
	RoleObserver CollaboratorRole = "observer"
)

type Collaborator struct {
	Actor string           `json:"actor"`
	Role  CollaboratorRole `json:"role"`
}

type CollaborationNote struct {
	EventID        string `json:"event_id"`
	RegistrationID string `json:"registration_id"`
	Actor          string `json:"actor"`
	Text           string `json:"text"`
}

var ErrInvalidCollaborator = errors.New("invalid collaborator")

func ValidateCollaborator(collaborator Collaborator) error {
	if strings.TrimSpace(collaborator.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidCollaborator)
	}
	switch collaborator.Role {
	case RoleOwner, RoleReviewer, RoleOperator, RoleObserver:
		return nil
	default:
		return fmt.Errorf("%w: role is unknown", ErrInvalidCollaborator)
	}
}

func CanPerform(role CollaboratorRole, action string) bool {
	action = strings.ToLower(strings.TrimSpace(action))
	if role == RoleOwner {
		return true
	}
	if role == RoleReviewer {
		return action == "review" || action == "search" || action == "export"
	}
	if role == RoleOperator {
		return action == "batch" || action == "search" || action == "export"
	}
	return action == "search"
}

func RoleLabel(role CollaboratorRole) string {
	switch role {
	case RoleOwner:
		return "Event owner"
	case RoleReviewer:
		return "Registration reviewer"
	case RoleOperator:
		return "Batch operator"
	default:
		return "Read-only observer"
	}
}

func NormalizeActor(actor string) string {
	return strings.ToLower(strings.TrimSpace(actor))
}

func UniqueActors(audits []AuditEvent) []string {
	seen := make(map[string]bool)
	actors := make([]string, 0)
	for _, audit := range audits {
		actor := NormalizeActor(audit.Actor)
		if actor == "" || seen[actor] {
			continue
		}
		seen[actor] = true
		actors = append(actors, actor)
	}
	sort.Strings(actors)
	return actors
}

func NewCollaborationNote(eventID, registrationID, actor, text string) (CollaborationNote, error) {
	note := CollaborationNote{EventID: strings.TrimSpace(eventID), RegistrationID: strings.TrimSpace(registrationID), Actor: NormalizeActor(actor), Text: strings.TrimSpace(text)}
	if note.EventID == "" || note.Actor == "" || note.Text == "" {
		return CollaborationNote{}, errors.New("collaboration note requires event, actor, and text")
	}
	if len([]rune(note.Text)) > 500 {
		return CollaborationNote{}, errors.New("collaboration note is too long")
	}
	return note, nil
}

func ActionCounts(audits []AuditEvent) map[string]int {
	counts := make(map[string]int)
	for _, audit := range audits {
		counts[audit.Action]++
	}
	return counts
}
