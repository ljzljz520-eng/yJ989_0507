package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrInvalidEvent        = errors.New("invalid event")
	ErrInvalidRegistration = errors.New("invalid registration")
	ErrEventNotOpen        = errors.New("event is not open")
	ErrCapacityReached     = errors.New("event capacity reached")
	ErrInvalidTransition   = errors.New("invalid event transition")
	ErrInvalidReview       = errors.New("invalid review")
)

func ValidateEvent(event Event) error {
	if strings.TrimSpace(event.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidEvent)
	}
	if strings.TrimSpace(event.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidEvent)
	}
	if event.Capacity < 1 {
		return fmt.Errorf("%w: capacity must be positive", ErrInvalidEvent)
	}
	if event.Deadline.Before(event.RegistrationOpensAt) {
		return fmt.Errorf("%w: deadline precedes opening", ErrInvalidEvent)
	}
	if event.Status == "" {
		return fmt.Errorf("%w: status is required", ErrInvalidEvent)
	}
	return nil
}

func ValidateRegistration(registration Registration) error {
	if strings.TrimSpace(registration.ID) == "" || strings.TrimSpace(registration.EventID) == "" {
		return fmt.Errorf("%w: identifiers are required", ErrInvalidRegistration)
	}
	if strings.TrimSpace(registration.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidRegistration)
	}
	if !ValidPhone(registration.Phone) {
		return fmt.Errorf("%w: phone is invalid", ErrInvalidRegistration)
	}
	if len([]rune(registration.Note)) > 200 {
		return fmt.Errorf("%w: note is too long", ErrInvalidRegistration)
	}
	return nil
}

func ValidPhone(phone string) bool {
	digits := 0
	for _, character := range phone {
		if unicode.IsDigit(character) {
			digits++
			continue
		}
		if character != '+' && character != '-' && character != ' ' && character != '(' && character != ')' {
			return false
		}
	}
	return digits >= 7 && digits <= 15
}

func NormalizePhone(phone string) string {
	var builder strings.Builder
	for _, character := range phone {
		if unicode.IsDigit(character) || (character == '+' && builder.Len() == 0) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

func CanAcceptRegistration(event Event, registrationCount int, nowIsBeforeDeadline bool) error {
	if event.Status != EventOpen {
		return ErrEventNotOpen
	}
	if !nowIsBeforeDeadline {
		return fmt.Errorf("%w: deadline passed", ErrEventNotOpen)
	}
	if registrationCount >= event.Capacity {
		return ErrCapacityReached
	}
	return nil
}

func TransitionEvent(event Event, target EventStatus) error {
	if event.Status == target {
		return nil
	}
	allowed := map[EventStatus][]EventStatus{
		EventDraft:    {EventOpen, EventArchived},
		EventOpen:     {EventClosed, EventArchived},
		EventClosed:   {EventOpen, EventArchived},
		EventArchived: {},
	}
	for _, candidate := range allowed[event.Status] {
		if candidate == target {
			return nil
		}
	}
	return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, event.Status, target)
}

func ValidateReview(decision ReviewDecision, reviewer string) error {
	if decision != DecisionApprove && decision != DecisionReject {
		return ErrInvalidReview
	}
	if strings.TrimSpace(reviewer) == "" {
		return fmt.Errorf("%w: reviewer is required", ErrInvalidReview)
	}
	return nil
}

func IsVisibleRegistration(status RegistrationStatus) bool {
	return status != RegistrationCancelled && status != RegistrationRejected
}
