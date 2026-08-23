package domain

import (
	"testing"
	"time"
)

func TestRankRegistrationsUsesSubmissionOrderForTies(t *testing.T) {
	base := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)
	items := []Registration{
		{ID: "later", Score: 50, SubmittedAt: base.Add(time.Minute)},
		{ID: "earlier", Score: 50, SubmittedAt: base},
		{ID: "higher", Score: 60, SubmittedAt: base.Add(2 * time.Minute)},
	}
	ranked := RankRegistrations(items)
	if ranked[0].ID != "higher" || ranked[1].ID != "earlier" || ranked[2].ID != "later" {
		t.Fatalf("unexpected ranking: %v", RegistrationIDs(ranked))
	}
	page, total := PageRegistrations(ranked, 2, 2)
	if total != 3 || len(page) != 1 || page[0].ID != "later" {
		t.Fatalf("unexpected page: total=%d page=%v", total, RegistrationIDs(page))
	}
}
