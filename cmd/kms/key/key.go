package key

import (
	"github.com/exoscale/cli/cmd/kms"
	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "KMS key",
}

func init() {
	kms.KMSCmd.AddCommand(keyCmd)
}
