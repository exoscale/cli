package apikey

import (
	"github.com/spf13/cobra"
)

// Cmd is the root command for apikey subcommands.
var Cmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage AI API keys",
}
