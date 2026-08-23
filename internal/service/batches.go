package service

import (
	"fmt"

	"activityregistration/internal/domain"
)

func (service *Service) CreateBatch(eventID, batchID string) (domain.Batch, error) {
	if _, err := service.store.GetEvent(eventID); err != nil {
		return domain.Batch{}, err
	}
	if batchID == "" {
		batchID = service.batchIDs.Next()
	}
	registrations, err := service.store.ListRegistrations(eventID)
	if err != nil {
		return domain.Batch{}, err
	}
	batch := domain.Batch{ID: batchID, EventID: eventID, Step: domain.BatchCreated, Total: len(registrations), UpdatedAt: service.clock.Now()}
	if err := service.store.PutBatch(batch); err != nil {
		return domain.Batch{}, err
	}
	if err := service.recordAudit(eventID, "", "batch.created", "admin", batch.ID); err != nil {
		return domain.Batch{}, err
	}
	return batch, nil
}

func (service *Service) GetBatch(eventID, batchID string) (domain.Batch, error) {
	return service.store.GetBatch(eventID, batchID)
}

func (service *Service) AdvanceBatch(eventID, batchID string) (domain.Batch, error) {
	batch, err := service.GetBatch(eventID, batchID)
	if err != nil {
		return domain.Batch{}, err
	}
	if batch.Step == domain.BatchArchived || batch.Step == domain.BatchCancelled {
		return domain.Batch{}, ErrBatchFinished
	}
	if batch.CancelRequested {
		batch.Step = domain.BatchCancelled
		batch.UpdatedAt = service.clock.Now()
		if err := service.store.PutBatch(batch); err != nil {
			return domain.Batch{}, err
		}
		return batch, service.recordAudit(eventID, "", "batch.cancelled", "admin", batch.ID)
	}
	switch batch.Step {
	case domain.BatchCreated:
		batch.Step = domain.BatchPrepared
	case domain.BatchPrepared:
		batch.Step = domain.BatchRanked
	case domain.BatchRanked:
		batch.Step = domain.BatchConfirmed
	case domain.BatchConfirmed:
		batch.Step = domain.BatchArchived
	default:
		return domain.Batch{}, fmt.Errorf("%w: %s", ErrBatchStep, batch.Step)
	}
	batch.Cursor++
	batch.UpdatedAt = service.clock.Now()
	if err := service.store.PutBatch(batch); err != nil {
		return domain.Batch{}, err
	}
	return batch, service.recordAudit(eventID, "", "batch.step."+string(batch.Step), "admin", batch.ID)
}

func (service *Service) RequestBatchCancellation(eventID, batchID, actor string) (domain.Batch, error) {
	batch, err := service.GetBatch(eventID, batchID)
	if err != nil {
		return domain.Batch{}, err
	}
	if batch.Step == domain.BatchArchived || batch.Step == domain.BatchCancelled {
		return domain.Batch{}, ErrBatchFinished
	}
	batch.CancelRequested = true
	batch.UpdatedAt = service.clock.Now()
	if err := service.store.PutBatch(batch); err != nil {
		return domain.Batch{}, err
	}
	return batch, service.recordAudit(eventID, "", "batch.cancel.requested", actor, batch.ID)
}

func (service *Service) ConfirmBatch(eventID, batchID, actor string) (domain.Batch, error) {
	batch, err := service.GetBatch(eventID, batchID)
	if err != nil {
		return domain.Batch{}, err
	}
	if batch.Step != domain.BatchRanked {
		return domain.Batch{}, fmt.Errorf("%w: confirmation requires ranked step", ErrBatchStep)
	}
	if batch.CancelRequested {
		batch.Step = domain.BatchCancelled
	} else {
		batch.Step = domain.BatchConfirmed
	}
	batch.Cursor++
	batch.UpdatedAt = service.clock.Now()
	if err := service.store.PutBatch(batch); err != nil {
		return domain.Batch{}, err
	}
	return batch, service.recordAudit(eventID, "", "batch.confirmed", actor, batch.ID)
}

func (service *Service) BatchProgress(eventID, batchID string) (string, error) {
	batch, err := service.GetBatch(eventID, batchID)
	if err != nil {
		return "", err
	}
	return string(batch.Step), nil
}
