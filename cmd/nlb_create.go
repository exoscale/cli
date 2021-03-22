package cmd

import (
	"fmt"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var nlbCreateCmd = &cobra.Command{
	Use:     "create NAME",
	Short:   "Create a Network Load Balancer",
	Aliases: gCreateAlias,
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

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		nlb, err := cs.CreateNetworkLoadBalancer(ctx, zone, &exov2.NetworkLoadBalancer{
			Name:        args[0],
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("unable to create Network Load Balancer: %s", err)
		}

		if !gQuiet {
			return output(showNLB(zone, nlb.ID))
		}

		return nil
	},
}

func init() {
	nlbCreateCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbCreateCmd.Flags().String("description", "", "service description")
	nlbCmd.AddCommand(nlbCreateCmd)
}
