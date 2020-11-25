package cmd

import (
	"errors"
	"fmt"

	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var sksDeleteCmd = &cobra.Command{
	Use:     "delete <name | ID>",
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

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, args[0])
		if err != nil {
			return err
		}

		if len(cluster.Nodepools) > 0 {
			nodepoolsRemaining := len(cluster.Nodepools)

			if deleteNodepools {
				for _, nodepool := range cluster.Nodepools {
					if !askQuestion(fmt.Sprintf("Do you really want to delete SKS cluster Nodepool %q?",
						nodepool.Name)) {
						continue
					}

					if err = cluster.DeleteNodepool(ctx, nodepool); err != nil {
						return fmt.Errorf("error deleting Nodepool: %s", err)
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
			if !askQuestion(fmt.Sprintf("Do you really want to delete SKS cluster %q?", args[0])) {
				return nil
			}
		}

		if err := cs.DeleteSKSCluster(ctx, zone, cluster.ID); err != nil {
			return fmt.Errorf("unable to delete SKS cluster: %s", err)
		}

		if !gQuiet {
			cmd.Println("SKS cluster deleted successfully")
		}

		return nil
	},
}

func init() {
	sksDeleteCmd.Flags().BoolP("force", "f", false,
		"Attempt to delete without prompting for confirmation")
	sksDeleteCmd.Flags().BoolP("nodepools", "n", false,
		"Delete existing Nodepools before deleting the SKS cluster")
	sksDeleteCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksCmd.AddCommand(sksDeleteCmd)
}
