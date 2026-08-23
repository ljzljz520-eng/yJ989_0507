package store

import "activityregistration/internal/domain"

func auditKey(eventID, auditID string) string {
	return eventID + "\x00" + auditID
}

func (store *Store) PutAuditEvent(event domain.AuditEvent) error {
	return store.put(bucketNames[3], auditKey(event.EventID, event.ID), event)
}

func (store *Store) ListAuditEvents(eventID string) ([]domain.AuditEvent, error) {
	items, err := list(store, bucketNames[3], func(data []byte) (domain.AuditEvent, error) {
		var event domain.AuditEvent
		return event, decode(data, &event)
	})
	if err != nil {
		return nil, err
	}
	filtered := items[:0]
	for _, event := range items {
		if event.EventID == eventID {
			filtered = append(filtered, event)
		}
	}
	return filtered, nil
}

func (store *Store) CountAuditEvents(eventID string) (int, error) {
	items, err := store.ListAuditEvents(eventID)
	return len(items), err
}
