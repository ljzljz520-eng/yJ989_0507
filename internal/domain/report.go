package domain

import (
	"sort"
	"strings"
	"time"
)

type EventSummary struct {
	EventID       string      `json:"event_id"`
	Name          string      `json:"name"`
	Status        EventStatus `json:"status"`
	Capacity      int         `json:"capacity"`
	Total         int         `json:"total"`
	Pending       int         `json:"pending"`
	Approved      int         `json:"approved"`
	Rejected      int         `json:"rejected"`
	Cancelled     int         `json:"cancelled"`
	Remaining     int         `json:"remaining"`
	LastUpdatedAt time.Time   `json:"last_updated_at"`
}

type ReviewerQueue struct {
	EventID       string         `json:"event_id"`
	Items         []Registration `json:"items"`
	PendingCount  int            `json:"pending_count"`
	ApprovedCount int            `json:"approved_count"`
}

func CapacityRemaining(event Event, registrations []Registration) int {
	approved := 0
	for _, registration := range registrations {
		if registration.Status == RegistrationApproved {
			approved++
		}
	}
	remaining := event.Capacity - approved
	if remaining < 0 {
		return 0
	}
	return remaining
}

func IsEventInWindow(event Event, now time.Time) bool {
	return event.Status == EventOpen && !now.Before(event.RegistrationOpensAt) && now.Before(event.Deadline)
}

func StatusCounts(registrations []Registration) map[RegistrationStatus]int {
	counts := map[RegistrationStatus]int{
		RegistrationPending:   0,
		RegistrationApproved:  0,
		RegistrationRejected:  0,
		RegistrationCancelled: 0,
	}
	for _, registration := range registrations {
		counts[registration.Status]++
	}
	return counts
}

func SortEventsByDeadline(events []Event) []Event {
	result := append([]Event(nil), events...)
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].Deadline.Equal(result[right].Deadline) {
			return result[left].ID < result[right].ID
		}
		return result[left].Deadline.Before(result[right].Deadline)
	})
	return result
}

func FilterEventsByStatus(events []Event, status EventStatus) []Event {
	filtered := make([]Event, 0, len(events))
	for _, event := range events {
		if event.Status == status {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func SearchText(value, query string) bool {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), needle)
}

func ValidatePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

func RegistrationIsEditable(registration Registration) bool {
	return registration.Status == RegistrationPending
}

func RegistrationDisplayName(registration Registration) string {
	name := strings.TrimSpace(registration.Name)
	if name == "" {
		return "Unnamed applicant"
	}
	return name
}

func MaskPhone(phone string) string {
	runes := []rune(phone)
	if len(runes) <= 4 {
		return strings.Repeat("*", len(runes))
	}
	return string(runes[:3]) + strings.Repeat("*", len(runes)-5) + string(runes[len(runes)-2:])
}

func HasCapacity(event Event, registrations []Registration) bool {
	return CapacityRemaining(event, registrations) > 0
}
