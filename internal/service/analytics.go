package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"activityregistration/internal/domain"
	"activityregistration/internal/store"
)

func (service *Service) EventSummary(eventID string) (domain.EventSummary, error) {
	event, err := service.store.GetEvent(eventID)
	if err != nil {
		return domain.EventSummary{}, err
	}
	registrations, err := service.store.ListRegistrations(eventID)
	if err != nil {
		return domain.EventSummary{}, err
	}
	counts := domain.StatusCounts(registrations)
	lastUpdated := event.UpdatedAt
	for _, registration := range registrations {
		if registration.UpdatedAt.After(lastUpdated) {
			lastUpdated = registration.UpdatedAt
		}
	}
	return domain.EventSummary{EventID: event.ID, Name: event.Name, Status: event.Status, Capacity: event.Capacity, Total: len(registrations), Pending: counts[domain.RegistrationPending], Approved: counts[domain.RegistrationApproved], Rejected: counts[domain.RegistrationRejected], Cancelled: counts[domain.RegistrationCancelled], Remaining: domain.CapacityRemaining(event, registrations), LastUpdatedAt: lastUpdated}, nil
}

func (service *Service) ReviewerQueue(eventID string, limit int) (domain.ReviewerQueue, error) {
	if limit < 1 {
		limit = 25
	}
	pending, err := service.store.ListRegistrationsByStatus(eventID, domain.RegistrationPending)
	if err != nil {
		return domain.ReviewerQueue{}, err
	}
	sort.SliceStable(pending, func(left, right int) bool {
		if pending[left].SubmittedAt.Equal(pending[right].SubmittedAt) {
			return pending[left].ID < pending[right].ID
		}
		return pending[left].SubmittedAt.Before(pending[right].SubmittedAt)
	})
	if len(pending) > limit {
		pending = pending[:limit]
	}
	approved, err := service.store.CountApproved(eventID)
	if err != nil {
		return domain.ReviewerQueue{}, err
	}
	return domain.ReviewerQueue{EventID: eventID, Items: pending, PendingCount: len(pending), ApprovedCount: approved}, nil
}

func (service *Service) UpdateRegistrationNote(eventID, registrationID, note, actor string) (domain.Registration, error) {
	registration, err := service.store.GetRegistration(eventID, registrationID)
	if err != nil {
		return domain.Registration{}, err
	}
	if !domain.RegistrationIsEditable(registration) {
		return domain.Registration{}, fmt.Errorf("registration %s cannot be edited", registrationID)
	}
	registration.Note = strings.TrimSpace(note)
	registration.Score = domain.ScoreRegistration(registration)
	registration.UpdatedAt = service.clock.Now()
	if err := domain.ValidateRegistration(registration); err != nil {
		return domain.Registration{}, err
	}
	if err := service.store.PutRegistration(registration); err != nil {
		return domain.Registration{}, err
	}
	if err := service.recordAudit(eventID, registrationID, "registration.note.updated", actor, registration.Note); err != nil {
		return domain.Registration{}, err
	}
	return registration, nil
}

func (service *Service) CancelRegistration(eventID, registrationID, actor string) (domain.Registration, error) {
	registration, err := service.store.GetRegistration(eventID, registrationID)
	if err != nil {
		return domain.Registration{}, err
	}
	if registration.Status == domain.RegistrationCancelled {
		return registration, nil
	}
	if registration.Status == domain.RegistrationRejected {
		return domain.Registration{}, errors.New("rejected registration cannot be cancelled")
	}
	registration.Status = domain.RegistrationCancelled
	registration.UpdatedAt = service.clock.Now()
	if err := service.store.PutRegistration(registration); err != nil {
		return domain.Registration{}, err
	}
	if err := service.recordAudit(eventID, registrationID, "registration.cancelled", actor, registration.ID); err != nil {
		return domain.Registration{}, err
	}
	return registration, nil
}

func (service *Service) RestoreRegistration(eventID, registrationID, actor string) (domain.Registration, error) {
	registration, err := service.store.GetRegistration(eventID, registrationID)
	if err != nil {
		return domain.Registration{}, err
	}
	if registration.Status != domain.RegistrationCancelled {
		return domain.Registration{}, errors.New("only cancelled registration can be restored")
	}
	registration.Status = domain.RegistrationPending
	registration.UpdatedAt = service.clock.Now()
	if err := service.store.PutRegistration(registration); err != nil {
		return domain.Registration{}, err
	}
	if err := service.recordAudit(eventID, registrationID, "registration.restored", actor, registration.ID); err != nil {
		return domain.Registration{}, err
	}
	return registration, nil
}

func (service *Service) SearchRegistrations(eventID, query string, page, pageSize int) (ListResult, error) {
	page, pageSize = domain.ValidatePage(page, pageSize)
	items, err := service.store.FindRegistrations(eventID, query)
	if err != nil {
		return ListResult{}, err
	}
	visible := make([]domain.Registration, 0, len(items))
	for _, registration := range items {
		if domain.IsVisibleRegistration(registration.Status) {
			visible = append(visible, registration)
		}
	}
	visible = domain.RankRegistrations(visible)
	pageItems, total := domain.PageRegistrations(visible, page, pageSize)
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return ListResult{Items: pageItems, Page: page, PageSize: pageSize, Total: total, TotalPages: pages}, nil
}

func (service *Service) ArchiveSummary(eventID string) (domain.EventSummary, error) {
	summary, err := service.EventSummary(eventID)
	if err != nil {
		return domain.EventSummary{}, err
	}
	if summary.Status != domain.EventArchived {
		return domain.EventSummary{}, fmt.Errorf("event %s is not archived", eventID)
	}
	return summary, nil
}

func (service *Service) ExportJobs(eventID string) ([]domain.ExportJob, error) {
	return service.store.ListExportJobs(eventID)
}

func (service *Service) BatchHistory(eventID string) ([]domain.Batch, error) {
	batches, err := service.store.ListBatches(eventID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(batches, func(left, right int) bool {
		if batches[left].UpdatedAt.Equal(batches[right].UpdatedAt) {
			return batches[left].ID < batches[right].ID
		}
		return batches[left].UpdatedAt.Before(batches[right].UpdatedAt)
	})
	return batches, nil
}

func (service *Service) DeleteEvent(eventID string) error {
	event, err := service.store.GetEvent(eventID)
	if err != nil {
		return err
	}
	if event.Status != domain.EventArchived {
		return fmt.Errorf("event %s must be archived before deletion", eventID)
	}
	return service.store.DeleteEvent(eventID)
}

func IsNotFound(err error) bool {
	return errors.Is(err, store.ErrNotFound)
}
