package cmd

import (
	fig "github.com/withfig/autocomplete-tools/integrations/cobra"
)

func init() {
	integrationsCmd.AddCommand(fig.CreateCompletionSpecCommand())
}
