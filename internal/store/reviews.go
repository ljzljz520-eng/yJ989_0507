package store

import (
	"activityregistration/internal/domain"
)

func reviewKey(eventID, reviewID string) string {
	return eventID + "\x00" + reviewID
}

func (store *Store) PutReview(review domain.ReviewRecord) error {
	return store.put(bucketNames[2], reviewKey(review.EventID, review.ID), review)
}

func (store *Store) GetReview(eventID, reviewID string) (domain.ReviewRecord, error) {
	var review domain.ReviewRecord
	err := store.get(bucketNames[2], reviewKey(eventID, reviewID), &review)
	return review, err
}

func (store *Store) ListReviews(eventID, registrationID string) ([]domain.ReviewRecord, error) {
	items, err := list(store, bucketNames[2], func(data []byte) (domain.ReviewRecord, error) {
		var review domain.ReviewRecord
		return review, decode(data, &review)
	})
	if err != nil {
		return nil, err
	}
	filtered := items[:0]
	for _, review := range items {
		if review.EventID == eventID && (registrationID == "" || review.RegistrationID == registrationID) {
			filtered = append(filtered, review)
		}
	}
	return filtered, nil
}

func (store *Store) LatestReview(eventID, registrationID string) (domain.ReviewRecord, error) {
	reviews, err := store.ListReviews(eventID, registrationID)
	if err != nil {
		return domain.ReviewRecord{}, err
	}
	if len(reviews) == 0 {
		return domain.ReviewRecord{}, ErrNotFound
	}
	latest := reviews[0]
	for _, review := range reviews[1:] {
		if review.ReviewedAt.After(latest.ReviewedAt) {
			latest = review
		}
	}
	return latest, nil
}
