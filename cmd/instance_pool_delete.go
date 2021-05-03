package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var instancePoolDeleteCmd = &cobra.Command{
	Use:     "delete NAME|ID",
	Short:   "Delete an Instance Pool",
	Aliases: gDeleteAlias,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		instancePool, err := lookupInstancePool(ctx, zone, args[0])
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete Instance Pool %q?", instancePool.Name)) {
				return nil
			}
		}

		// Ensure the Instance Pool is not attached to a NLB service.
		nlbs, err := cs.ListNetworkLoadBalancers(gContext, zone)
		if err != nil {
			return fmt.Errorf("unable to list Network Load Balancers: %v", err)
		}

		for _, nlb := range nlbs {
			for _, svc := range nlb.Services {
				if svc.InstancePoolID == instancePool.ID {
					return fmt.Errorf("Instance Pool %q is still referenced by NLB service %s/%s", // nolint
						instancePool.Name, nlb.Name, svc.Name)
				}
			}
		}

		decorateAsyncOperation(fmt.Sprintf("Deleting Instance Pool %q...", instancePool.Name), func() {
			err = cs.DeleteInstancePool(ctx, zone, instancePool.ID)
		})
		if err != nil {
			return err
		}

		if !gQuiet {
			cmd.Println("Instance Pool deleted successfully")
		}

		return nil
	},
}

func init() {
	instancePoolDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	instancePoolDeleteCmd.Flags().StringP("zone", "z", "", "Instance Pool zone")
	instancePoolCmd.AddCommand(instancePoolDeleteCmd)
}
