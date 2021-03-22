package cmd

import (
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksDeleteCmd = &cobra.Command{
	Use:     "delete NAME|ID",
	Short:   "Delete a SKS cluster",
	Aliases: gRemoveAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		deleteNodepools, err := cmd.Flags().GetBool("nodepools")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, args[0])
		if err != nil {
			return err
		}

		if len(cluster.Nodepools) > 0 {
			nodepoolsRemaining := len(cluster.Nodepools)

			if deleteNodepools {
				for _, nodepool := range cluster.Nodepools {
					nodepool := nodepool

					if !force {
						if !askQuestion(fmt.Sprintf("Are you sure you want to delete Nodepool %q?",
							nodepool.Name)) {
							continue
						}
					}

					decorateAsyncOperation(fmt.Sprintf("Deleting Nodepool %q...", nodepool.Name), func() {
						err = cluster.DeleteNodepool(ctx, nodepool)
					})
					if err != nil {
						return err
					}
					nodepoolsRemaining--
				}
			}

			// It's not possible to delete a SKS cluster that still has Nodepools, no need to go further.
			if nodepoolsRemaining > 0 {
				return errors.New("impossible to delete the SKS cluster: Nodepools still present")
			}
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete SKS cluster %q?", args[0])) {
				return nil
			}
		}

		decorateAsyncOperation(fmt.Sprintf("Deleting SKS cluster %q...", cluster.Name), func() {
			err = cs.DeleteSKSCluster(ctx, zone, cluster.ID)
		})
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	sksDeleteCmd.Flags().BoolP("force", "f", false,
		cmdFlagForceHelp)
	sksDeleteCmd.Flags().BoolP("nodepools", "n", false,
		"Delete existing Nodepools before deleting the SKS cluster")
	sksDeleteCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksCmd.AddCommand(sksDeleteCmd)
}
