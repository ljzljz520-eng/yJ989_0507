package service

import (
	"fmt"
	"io"

	"activityregistration/internal/importer"
)

func (service *Service) ImportRegistrations(eventID string, reader io.Reader) (ImportReport, error) {
	result, err := importer.ReadRegistrations(reader)
	if err != nil {
		return ImportReport{}, err
	}
	report := ImportReport{Invalid: append([]string(nil), result.Invalid...)}
	for _, row := range result.Rows {
		_, submitErr := service.SubmitRegistration(eventID, RegistrationInput{Name: row.Name, Phone: row.Phone, Note: row.Note})
		if submitErr != nil {
			report.Invalid = append(report.Invalid, fmt.Sprintf("%s: %v", row.Name, submitErr))
			continue
		}
		report.Imported++
	}
	return report, nil
}
