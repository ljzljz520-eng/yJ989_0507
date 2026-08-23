package store

import (
	"fmt"
	"strings"

	"activityregistration/internal/domain"
)

func registrationKey(eventID, registrationID string) string {
	return eventID + "\x00" + registrationID
}

func (store *Store) PutRegistration(registration domain.Registration) error {
	return store.put(bucketNames[1], registrationKey(registration.EventID, registration.ID), registration)
}

func (store *Store) GetRegistration(eventID, registrationID string) (domain.Registration, error) {
	var registration domain.Registration
	err := store.get(bucketNames[1], registrationKey(eventID, registrationID), &registration)
	return registration, err
}

func (store *Store) DeleteRegistration(eventID, registrationID string) error {
	return store.delete(bucketNames[1], registrationKey(eventID, registrationID))
}

func (store *Store) ListRegistrations(eventID string) ([]domain.Registration, error) {
	items, err := list(store, bucketNames[1], func(data []byte) (domain.Registration, error) {
		var registration domain.Registration
		return registration, decode(data, &registration)
	})
	if err != nil {
		return nil, err
	}
	filtered := items[:0]
	for _, registration := range items {
		if registration.EventID == eventID {
			filtered = append(filtered, registration)
		}
	}
	return filtered, nil
}

func (store *Store) CountRegistrations(eventID string) (int, error) {
	registrations, err := store.ListRegistrations(eventID)
	return len(registrations), err
}

func (store *Store) FindRegistrationByPhone(eventID, phone string) (domain.Registration, error) {
	registrations, err := store.ListRegistrations(eventID)
	if err != nil {
		return domain.Registration{}, err
	}
	for _, registration := range registrations {
		if strings.TrimSpace(registration.Phone) == strings.TrimSpace(phone) {
			return registration, nil
		}
	}
	return domain.Registration{}, fmt.Errorf("phone %s: %w", phone, ErrNotFound)
}

func (store *Store) UpdateRegistrationStatus(eventID, registrationID string, status domain.RegistrationStatus, reviewer string) (domain.Registration, error) {
	registration, err := store.GetRegistration(eventID, registrationID)
	if err != nil {
		return domain.Registration{}, err
	}
	registration.Status = status
	registration.ReviewedBy = reviewer
	if err := store.PutRegistration(registration); err != nil {
		return domain.Registration{}, err
	}
	return registration, nil
}
