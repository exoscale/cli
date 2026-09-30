package model

import (
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

// Cmd is the root command for model subcommands.
var Cmd = &cobra.Command{
	Use:   "model",
	Short: "Manage AI models",
}

func validateVisibility(v v3.ListModelsResponseEntryVisibility) error {
	if v != "" &&
		v != v3.ListModelsResponseEntryVisibilityPublic &&
		v != v3.ListModelsResponseEntryVisibilityPrivate {
		return fmt.Errorf("invalid --visibility %q: must be 'public' or 'private'", v)
	}
	return nil
}
