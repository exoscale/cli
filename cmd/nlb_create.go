package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var nlbCreateCmd = &cobra.Command{
	Use:     "create <name>",
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

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
		nlb, err := cs.CreateNetworkLoadBalancer(ctx, zone, &egoscale.NetworkLoadBalancer{
			Name:        args[0],
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("unable to create Network Load Balancer: %s", err)
		}

		if !gQuiet {
			return output(showNLB(nlb.ID, zone))
		}

		return nil
	},
}

func init() {
	nlbCreateCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbCreateCmd.Flags().String("description", "", "service description")
	nlbCmd.AddCommand(nlbCreateCmd)
}
