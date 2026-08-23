package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const InputTimeLayout = time.RFC3339

func ParseInputTime(value string) (time.Time, error) {
	parsed, err := time.Parse(InputTimeLayout, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("parse input time: %w", err)
	}
	return parsed.UTC(), nil
}

func FormatInputTime(value time.Time) string {
	return value.UTC().Format(InputTimeLayout)
}

func ParseCapacity(value string) (int, error) {
	capacity, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || capacity < 1 {
		return 0, fmt.Errorf("capacity must be a positive integer")
	}
	return capacity, nil
}

func ParseEventStatus(value string) (EventStatus, error) {
	status := EventStatus(strings.ToLower(strings.TrimSpace(value)))
	switch status {
	case EventDraft, EventOpen, EventClosed, EventArchived:
		return status, nil
	default:
		return "", fmt.Errorf("unknown event status %q", value)
	}
}

func ParseReviewDecision(value string) (ReviewDecision, error) {
	decision := ReviewDecision(strings.ToLower(strings.TrimSpace(value)))
	if decision != DecisionApprove && decision != DecisionReject {
		return "", fmt.Errorf("unknown review decision %q", value)
	}
	return decision, nil
}

func ParseBatchStep(value string) (BatchStep, error) {
	step := BatchStep(strings.ToLower(strings.TrimSpace(value)))
	switch step {
	case BatchCreated, BatchPrepared, BatchRanked, BatchConfirmed, BatchArchived, BatchCancelled:
		return step, nil
	default:
		return "", fmt.Errorf("unknown batch step %q", value)
	}
}
