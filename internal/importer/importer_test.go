package importer

import (
	"strings"
	"testing"
)

func TestReadRegistrationsValidatesRows(t *testing.T) {
	input := "name,phone,note\nAda,13800000000,access\nBad,invalid,missing\n"
	result, err := ReadRegistrations(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 || result.Rows[0].Name != "Ada" || InvalidCount(result) != 1 {
		t.Fatalf("unexpected import result: %+v", result)
	}
	if _, err := ReadRegistrations(strings.NewReader("wrong,header\n")); err == nil {
		t.Fatal("invalid header should fail")
	}
}
