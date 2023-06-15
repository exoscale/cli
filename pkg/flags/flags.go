package flags

import (
	"github.com/spf13/cobra"
)

const (
	Versions = "versions"
)

func AddVersionsFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(Versions, false, "list all versions of objects(if the bucket is versioned)")
}
