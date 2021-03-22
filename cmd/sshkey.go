package cmd

import (
	"github.com/spf13/cobra"
)

var sshkeyCmd = &cobra.Command{
	Use:   "sshkey",
	Short: "SSH key pairs management",
}

func init() {
	RootCmd.AddCommand(sshkeyCmd)
}
