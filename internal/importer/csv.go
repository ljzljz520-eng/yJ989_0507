package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"activityregistration/internal/domain"
)

type Row struct {
	Name  string
	Phone string
	Note  string
}

type Result struct {
	Rows    []Row
	Invalid []string
}

func ReadRegistrations(reader io.Reader) (Result, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	header, err := csvReader.Read()
	if err != nil {
		return Result{}, fmt.Errorf("read import header: %w", err)
	}
	if !sameHeader(header) {
		return Result{}, errorsForHeader(header)
	}
	result := Result{Rows: make([]Row, 0)}
	line := 1
	for {
		record, readErr := csvReader.Read()
		if readErr == io.EOF {
			break
		}
		line++
		if readErr != nil {
			result.Invalid = append(result.Invalid, fmt.Sprintf("line %d: %v", line, readErr))
			continue
		}
		row, validationErr := rowFromRecord(record)
		if validationErr != nil {
			result.Invalid = append(result.Invalid, fmt.Sprintf("line %d: %v", line, validationErr))
			continue
		}
		result.Rows = append(result.Rows, row)
	}
	return result, nil
}

func sameHeader(header []string) bool {
	wanted := []string{"name", "phone", "note"}
	if len(header) != len(wanted) {
		return false
	}
	for index, value := range wanted {
		if strings.ToLower(strings.TrimSpace(header[index])) != value {
			return false
		}
	}
	return true
}

func errorsForHeader(header []string) error {
	return fmt.Errorf("expected name, phone, note header, got %v", header)
}

func rowFromRecord(record []string) (Row, error) {
	if len(record) != 3 {
		return Row{}, errorsForHeader(record)
	}
	row := Row{Name: strings.TrimSpace(record[0]), Phone: strings.TrimSpace(record[1]), Note: strings.TrimSpace(record[2])}
	if err := domain.ValidateRegistration(domain.Registration{ID: "import", EventID: "event", Name: row.Name, Phone: row.Phone, Note: row.Note}); err != nil {
		return Row{}, err
	}
	return row, nil
}

func ValidRows(result Result) []Row {
	return append([]Row(nil), result.Rows...)
}

func InvalidCount(result Result) int {
	return len(result.Invalid)
}
