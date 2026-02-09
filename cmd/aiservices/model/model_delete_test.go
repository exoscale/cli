package model

import (
	"testing"

	exocmd "github.com/exoscale/cli/cmd"
	v3 "github.com/exoscale/egoscale/v3"
)

func TestModelDelete(t *testing.T) {
	ts := newModelTestServer(t)
	defer modelSetup(t, ts)()
	ts.models = []v3.ListModelsResponseEntry{
		{ID: v3.UUID("11111111-1111-1111-1111-111111111111"), Name: "m1"},
		{ID: v3.UUID("22222222-2222-2222-2222-222222222222"), Name: "m2"},
	}

	// Not found without force
	cmd := &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Models: []string{"not-found"}, Force: false}
	if err := cmd.CmdRun(nil, nil); err == nil {
		t.Fatal("expected error for not found model without force")
	}
	// Not found with force (should skip with warning, no error)
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Models: []string{"not-found"}, Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("expected no error with force flag for not found model, got %v", err)
	}
	// Success by ID
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Models: []string{"11111111-1111-1111-1111-111111111111"}, Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model delete by ID: %v", err)
	}
	// Success by Name
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Models: []string{"m2"}, Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model delete by name: %v", err)
	}
	// multiple models
	cmd = &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings(), Models: []string{"m1", "22222222-2222-2222-2222-222222222222"}, Force: true}
	if err := cmd.CmdRun(nil, nil); err != nil {
		t.Fatalf("model delete multiple: %v", err)
	}
}

func TestModelDeleteCmd_CmdAliases(t *testing.T) {
	cmd := &ModelDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}
	aliases := cmd.CmdAliases()
	if len(aliases) == 0 {
		t.Fatal("CmdAliases() returned empty slice")
	}
	// Verify it returns the standard delete aliases
	expectedAliases := exocmd.GDeleteAlias
	if len(aliases) != len(expectedAliases) {
		t.Fatalf("expected %d aliases, got %d", len(expectedAliases), len(aliases))
	}
	for i, alias := range aliases {
		if alias != expectedAliases[i] {
			t.Fatalf("expected alias[%d] to be %q, got %q", i, expectedAliases[i], alias)
		}
	}
}
