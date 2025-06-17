package storage

import (
	"github.com/spf13/cobra"
)

var storageMetadataCmd = &cobra.Command{
	Use:     "metadata",
	Aliases: []string{"meta"},
	Short:   "Manage objects metadata",
}

func init() {
	storageCmd.AddCommand(storageMetadataCmd)
}
