package cmd

import (
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksNodepoolDeleteCmd = &cobra.Command{
	Use:     "delete <cluster name | ID> <Nodepool name | ID>",
	Short:   "Delete a SKS cluster Nodepool",
	Aliases: gRemoveAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c  = args[0]
			np = args[1]
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Do you really want to delete Nodepool %q?", args[1])) {
				return nil
			}
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, c)
		if err != nil {
			return err
		}

		for _, n := range cluster.Nodepools {
			if n.ID == np || n.Name == np {
				n := n
				decorateAsyncOperation(fmt.Sprintf("Deleting Nodepool %q...", np), func() {
					err = cluster.DeleteNodepool(ctx, n)
				})
				if err != nil {
					return err
				}

				return nil
			}
		}

		return errors.New("Nodepool not found") // nolint:golint
	},
}

func init() {
	sksNodepoolDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to delete without prompting for confirmation")
	sksNodepoolDeleteCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolCmd.AddCommand(sksNodepoolDeleteCmd)
}
