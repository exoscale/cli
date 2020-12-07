package cmd

import (
	"fmt"

	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var sksUpdateCmd = &cobra.Command{
	Use:   "update <name | ID>",
	Short: "Update a SKS cluster",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := args[0]

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, c)
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("name") {
			cluster.Name = name
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("description") {
			cluster.Description = description
		}

		decorateAsyncOperation(fmt.Sprintf("Updating SKS cluster %q...", c), func() {
			err = cs.UpdateSKSCluster(ctx, zone, cluster)
		})
		if err != nil {
			return err
		}

		if !gQuiet {
			return output(showSKSCluster(zone, cluster.ID))
		}

		return nil
	},
}

func init() {
	sksUpdateCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksUpdateCmd.Flags().String("name", "", "name")
	sksUpdateCmd.Flags().String("description", "", "description")
	sksCmd.AddCommand(sksUpdateCmd)
}
