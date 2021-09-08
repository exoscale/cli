package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var sshkeyCmd = &cobra.Command{
	Use:   "sshkey",
	Short: "SSH key pairs management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo sshkey" commands are deprecated and will be removed in a future
version, please use "exo compute ssh-key" replacement commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

func init() {
	RootCmd.AddCommand(sshkeyCmd)
}
