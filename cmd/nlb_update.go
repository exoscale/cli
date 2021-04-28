package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var nlbUpdateCmd = &cobra.Command{
	Use:   "update NAME|ID",
	Short: "Update a Network Load Balancer",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "missing arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		nlb, err := lookupNLB(ctx, zone, args[0])
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("name") {
			nlb.Name = name
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("description") {
			nlb.Description = description
		}

		if err := cs.UpdateNetworkLoadBalancer(ctx, zone, nlb); err != nil {
			return fmt.Errorf("unable to update Network Load Balancer: %s", err)
		}

		if !gQuiet {
			return output(showNLB(zone, nlb.ID))
		}

		return nil
	},
}

func init() {
	nlbUpdateCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbUpdateCmd.Flags().String("name", "", "service name")
	nlbUpdateCmd.Flags().String("description", "", "service description")
	nlbCmd.AddCommand(nlbUpdateCmd)
}
