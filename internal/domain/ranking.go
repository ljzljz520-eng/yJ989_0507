package domain

import (
	"sort"
	"strings"
)

func ScoreRegistration(registration Registration) int {
	score := 0
	if len([]rune(registration.Name)) >= 2 {
		score += 20
	}
	if ValidPhone(registration.Phone) {
		score += 30
	}
	if strings.TrimSpace(registration.Note) != "" {
		score += 10
	}
	if strings.Contains(strings.ToLower(registration.Note), "access") {
		score += 5
	}
	return score
}

func RankRegistrations(registrations []Registration) []Registration {
	result := append([]Registration(nil), registrations...)
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].Score != result[right].Score {
			return result[left].Score > result[right].Score
		}
		if !result[left].SubmittedAt.Equal(result[right].SubmittedAt) {
			return result[left].SubmittedAt.Before(result[right].SubmittedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result
}

func PageRegistrations(registrations []Registration, page, pageSize int) ([]Registration, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	start := (page - 1) * pageSize
	if start >= len(registrations) {
		return []Registration{}, len(registrations)
	}
	end := start + pageSize
	if end > len(registrations) {
		end = len(registrations)
	}
	return registrations[start:end], len(registrations)
}

func RegistrationIDs(registrations []Registration) []string {
	ids := make([]string, 0, len(registrations))
	for _, registration := range registrations {
		ids = append(ids, registration.ID)
	}
	return ids
}
