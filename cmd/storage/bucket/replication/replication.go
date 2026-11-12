package replication

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

var Cmd = &cobra.Command{
	Use:   "replication",
	Short: "Object Storage Bucket replication management",
	Long:  "Object Storage Bucket replication management",
}

// checkArgsAndZone keeps the behaviour of the legacy cobra commands: the
// exact number of positional arguments is enforced with cobra's error
// messages (RegisterCLICommand doesn't set cobra.Command.Args), and --zone
// defaults to the account zone and is required.
func checkArgsAndZone(cmd *cobra.Command, args []string, nArgs int) error {
	if err := cobra.ExactArgs(nArgs)(cmd, args); err != nil {
		return err
	}

	clicmd.CmdSetZoneFlagFromDefault(cmd)
	return clicmd.CmdCheckRequiredFlags(cmd, []string{"zone"})
}
