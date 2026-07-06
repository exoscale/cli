package crypto

import (
	"github.com/exoscale/cli/cmd/kms"
	"github.com/spf13/cobra"
)

var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "KMS key cryptographic operations",
}

func init() {
	kms.KMSCmd.AddCommand(cryptoCmd)
}
