package keystore

import (
	"context"

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

// resolveKeyStoreID returns the ID of the key store matching nameOrID.
func resolveKeyStoreID(ctx context.Context, client *v3.Client, nameOrID string) (v3.UUID, error) {
	list, err := client.ListKeyStores(ctx)
	if err != nil {
		return "", err
	}

	entry, err := list.FindListKeyStoresResponseEntry(nameOrID)
	if err != nil {
		return "", err
	}

	return entry.ID, nil
}
