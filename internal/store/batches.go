package store

import "activityregistration/internal/domain"

func batchKey(eventID, batchID string) string {
	return eventID + "\x00" + batchID
}

func (store *Store) PutBatch(batch domain.Batch) error {
	return store.put(bucketNames[4], batchKey(batch.EventID, batch.ID), batch)
}

func (store *Store) GetBatch(eventID, batchID string) (domain.Batch, error) {
	var batch domain.Batch
	err := store.get(bucketNames[4], batchKey(eventID, batchID), &batch)
	return batch, err
}

func (store *Store) ListBatches(eventID string) ([]domain.Batch, error) {
	items, err := list(store, bucketNames[4], func(data []byte) (domain.Batch, error) {
		var batch domain.Batch
		return batch, decode(data, &batch)
	})
	if err != nil {
		return nil, err
	}
	filtered := items[:0]
	for _, batch := range items {
		if batch.EventID == eventID {
			filtered = append(filtered, batch)
		}
	}
	return filtered, nil
}

func (store *Store) PutExportJob(job domain.ExportJob) error {
	return store.put(bucketNames[5], job.EventID+"\x00"+job.ID, job)
}

func (store *Store) ListExportJobs(eventID string) ([]domain.ExportJob, error) {
	items, err := list(store, bucketNames[5], func(data []byte) (domain.ExportJob, error) {
		var job domain.ExportJob
		return job, decode(data, &job)
	})
	if err != nil {
		return nil, err
	}
	filtered := items[:0]
	for _, job := range items {
		if job.EventID == eventID {
			filtered = append(filtered, job)
		}
	}
	return filtered, nil
}
