package crypto

import (
	"github.com/exoscale/cli/cmd/kms"
	"github.com/spf13/cobra"
)

// binaryInputHelp documents the input forms accepted by exocmd.BinaryValue arguments and flags.
const binaryInputHelp = `Binary values (plaintext, ciphertext, message, signature, encryption context) can be passed as base64, as file://<path> (file holding base64) or as fileb://<path> (raw binary file).`

var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "KMS key cryptographic operations",
}

func init() {
	kms.KMSCmd.AddCommand(cryptoCmd)
}
