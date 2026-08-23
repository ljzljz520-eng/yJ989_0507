package store

import (
	"sort"
	"strings"

	"activityregistration/internal/domain"
)

func (store *Store) ListRegistrationsByStatus(eventID string, status domain.RegistrationStatus) ([]domain.Registration, error) {
	registrations, err := store.ListRegistrations(eventID)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.Registration, 0)
	for _, registration := range registrations {
		if registration.Status == status {
			filtered = append(filtered, registration)
		}
	}
	return filtered, nil
}

func (store *Store) FindRegistrations(eventID, query string) ([]domain.Registration, error) {
	registrations, err := store.ListRegistrations(eventID)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return registrations, nil
	}
	filtered := make([]domain.Registration, 0)
	for _, registration := range registrations {
		if strings.Contains(strings.ToLower(registration.Name), needle) || strings.Contains(registration.Phone, needle) || strings.Contains(strings.ToLower(registration.Note), needle) {
			filtered = append(filtered, registration)
		}
	}
	return filtered, nil
}

func (store *Store) RegistrationStatusCounts(eventID string) (map[domain.RegistrationStatus]int, error) {
	registrations, err := store.ListRegistrations(eventID)
	if err != nil {
		return nil, err
	}
	return domain.StatusCounts(registrations), nil
}

func (store *Store) LatestAudit(eventID string) (domain.AuditEvent, error) {
	audits, err := store.ListAuditEvents(eventID)
	if err != nil {
		return domain.AuditEvent{}, err
	}
	if len(audits) == 0 {
		return domain.AuditEvent{}, ErrNotFound
	}
	latest := audits[0]
	for _, audit := range audits[1:] {
		if audit.OccurredAt.After(latest.OccurredAt) || (audit.OccurredAt.Equal(latest.OccurredAt) && audit.ID > latest.ID) {
			latest = audit
		}
	}
	return latest, nil
}

func (store *Store) SortReviewsByTime(reviews []domain.ReviewRecord) []domain.ReviewRecord {
	result := append([]domain.ReviewRecord(nil), reviews...)
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].ReviewedAt.Equal(result[right].ReviewedAt) {
			return result[left].ID < result[right].ID
		}
		return result[left].ReviewedAt.Before(result[right].ReviewedAt)
	})
	return result
}

func (store *Store) CountApproved(eventID string) (int, error) {
	registrations, err := store.ListRegistrationsByStatus(eventID, domain.RegistrationApproved)
	return len(registrations), err
}

func (store *Store) CountPending(eventID string) (int, error) {
	registrations, err := store.ListRegistrationsByStatus(eventID, domain.RegistrationPending)
	return len(registrations), err
}

func (store *Store) DeleteReview(eventID, reviewID string) error {
	return store.delete(bucketNames[2], reviewKey(eventID, reviewID))
}
