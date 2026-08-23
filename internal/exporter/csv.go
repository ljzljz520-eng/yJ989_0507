package exporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"activityregistration/internal/domain"
)

var columns = []string{"id", "event_id", "name", "phone", "note", "score", "submitted_at", "status", "reviewed_by"}

func Columns() []string {
	return append([]string(nil), columns...)
}

func WriteRegistrations(writer io.Writer, registrations []domain.Registration) error {
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write(columns); err != nil {
		return err
	}
	for _, registration := range registrations {
		row := []string{
			registration.ID,
			registration.EventID,
			registration.Name,
			registration.Phone,
			registration.Note,
			strconv.Itoa(registration.Score),
			domain.FormatInputTime(registration.SubmittedAt),
			string(registration.Status),
			registration.ReviewedBy,
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

func RenderRegistrations(registrations []domain.Registration) (string, error) {
	var builder strings.Builder
	if err := WriteRegistrations(&builder, registrations); err != nil {
		return "", fmt.Errorf("render registrations: %w", err)
	}
	return builder.String(), nil
}

func Summary(registrations []domain.Registration) map[string]int {
	result := map[string]int{"total": len(registrations)}
	for _, registration := range registrations {
		result[string(registration.Status)]++
	}
	return result
}

func FilterStatus(registrations []domain.Registration, status domain.RegistrationStatus) []domain.Registration {
	filtered := make([]domain.Registration, 0)
	for _, registration := range registrations {
		if registration.Status == status {
			filtered = append(filtered, registration)
		}
	}
	return filtered
}
