package httpapi

import (
	"activityregistration/internal/clock"
	"activityregistration/internal/service"
)

type eventPayload struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	RegistrationOpensAt string `json:"registration_opens_at"`
	Deadline            string `json:"deadline"`
	Capacity            int    `json:"capacity"`
}

func (payload eventPayload) toService() service.EventInput {
	opening, _ := clock.At(payload.RegistrationOpensAt)
	deadline, _ := clock.At(payload.Deadline)
	return service.EventInput{ID: payload.ID, Name: payload.Name, RegistrationOpensAt: opening, Deadline: deadline, Capacity: payload.Capacity}
}

type registrationPayload struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Note  string `json:"note"`
}

type reviewPayload struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
	Reviewer string `json:"reviewer"`
}
