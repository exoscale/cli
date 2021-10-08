package oapi

import (
	"testing"

	"github.com/gofrs/uuid"
)

func testRandomID(t *testing.T) string {
	id, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("unable to generate a new UUID: %s", err)
	}
	return id.String()
}
