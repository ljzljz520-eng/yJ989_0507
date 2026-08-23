package service

import (
	"fmt"
	"sort"

	"activityregistration/internal/domain"
	"activityregistration/internal/exporter"
)

func (service *Service) rankedRegistrations(eventID string, status domain.RegistrationStatus) ([]domain.Registration, error) {
	registrations, err := service.store.ListRegistrations(eventID)
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.Registration, 0, len(registrations))
	for _, registration := range registrations {
		if status != "" && registration.Status != status {
			continue
		}
		if status == "" && !domain.IsVisibleRegistration(registration.Status) {
			continue
		}
		filtered = append(filtered, registration)
	}
	sort.Slice(filtered, func(left, right int) bool {
		return filtered[left].Score > filtered[right].Score
	})
	return filtered, nil
}

func (service *Service) ListRegistrations(eventID string, options ListOptions) (ListResult, error) {
	if options.Page < 1 {
		options.Page = 1
	}
	if options.PageSize < 1 {
		options.PageSize = 25
	}
	registrations, err := service.rankedRegistrations(eventID, options.Status)
	if err != nil {
		return ListResult{}, err
	}
	items, total := domain.PageRegistrations(registrations, options.Page, options.PageSize)
	pages := total / options.PageSize
	if total%options.PageSize != 0 {
		pages++
	}
	return ListResult{Items: items, Page: options.Page, PageSize: options.PageSize, Total: total, TotalPages: pages}, nil
}

func (service *Service) CountVisible(eventID string) (int, error) {
	registrations, err := service.rankedRegistrations(eventID, "")
	return len(registrations), err
}

func (service *Service) ExportRegistrations(eventID, actor string) (string, domain.ExportJob, error) {
	registrations, err := service.rankedRegistrations(eventID, "")
	if err != nil {
		return "", domain.ExportJob{}, err
	}
	content, err := exporter.RenderRegistrations(registrations)
	if err != nil {
		return "", domain.ExportJob{}, err
	}
	job := domain.ExportJob{ID: service.exportIDs.Next(), EventID: eventID, RequestedBy: actor, Format: "csv", RowCount: len(registrations), CreatedAt: service.clock.Now()}
	if err := service.store.PutExportJob(job); err != nil {
		return "", domain.ExportJob{}, err
	}
	if err := service.recordAudit(eventID, "", "registration.exported", actor, fmt.Sprintf("rows=%d", len(registrations))); err != nil {
		return "", domain.ExportJob{}, err
	}
	return content, job, nil
}

func (service *Service) ExportSummary(eventID string) (map[string]int, error) {
	registrations, err := service.rankedRegistrations(eventID, "")
	if err != nil {
		return nil, err
	}
	return exporter.Summary(registrations), nil
}
