package storage

import (
	"github.com/spf13/cobra"
)

var storageHeaderCmd = &cobra.Command{
	Use:   "headers",
	Short: "Manage objects HTTP headers",
}

func init() {
	storageCmd.AddCommand(storageHeaderCmd)
}
