package flags

import (
	"github.com/spf13/cobra"
)

const (
	Versions              = "versions"
	ExcludeCurrentVersion = "exclude-current-version"
)

func AddVersionsFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(Versions, false, "TODO")
	cmd.Flags().Bool(ExcludeCurrentVersion, false, "list all versions except the current latest one")
}
