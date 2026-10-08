package keystore

import (
	"fmt"

	"github.com/exoscale/cli/cmd/kms"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "keystore",
	Short: "External key store (XKS) management",
}

func init() {
	kms.KMSCmd.AddCommand(Cmd)
}

// parseKeyStoreID validates a key store ID. Key stores are referenced by ID only.
func parseKeyStoreID(id string) (v3.UUID, error) {
	uuid, err := v3.ParseUUID(id)
	if err != nil {
		return "", fmt.Errorf("invalid key store ID %q: %w", id, err)
	}
	return uuid, nil
}
