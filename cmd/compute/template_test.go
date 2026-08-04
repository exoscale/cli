package compute

import (
	"context"
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
)

func TestResolveTemplateIDPassesUUIDThrough(t *testing.T) {
	const id = "5b473841-fd33-4b88-bcfe-a776abf15034"

	got, err := ResolveTemplateID(context.Background(), nil, id, "public", "ch-gva-2")
	if err != nil {
		t.Fatalf("ResolveTemplateID() error = %v", err)
	}
	if got != v3.UUID(id) {
		t.Errorf("ResolveTemplateID() = %q, want %q", got, id)
	}
}
