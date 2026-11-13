package metadata

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "metadata",
	Aliases: []string{"meta"},
	Short:   "Manage objects metadata",
}
