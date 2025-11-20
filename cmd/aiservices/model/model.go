package model

import (
	"github.com/spf13/cobra"
)

// Cmd is the root command for model subcommands.
var Cmd = &cobra.Command{
	Use:   "model",
	Short: "Manage AI models",
}
