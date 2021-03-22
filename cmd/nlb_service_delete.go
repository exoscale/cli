package cmd

import (
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var nlbServiceDeleteCmd = &cobra.Command{
	Use:     "delete NLB-NAME|ID SERVICE-NAME|ID",
	Short:   "Delete a Network Load Balancer service",
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
			nlbRef = args[0]
			svcRef = args[1]
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
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete Network Load Balancer service %q?", args[1])) {
				return nil
			}
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		nlb, err := lookupNLB(ctx, zone, nlbRef)
		if err != nil {
			return err
		}

		for _, svc := range nlb.Services {
			if svc.ID == svcRef || svc.Name == svcRef {
				if err := nlb.DeleteService(ctx, svc); err != nil {
					return err
				}

				if !gQuiet {
					cmd.Println("Service deleted successfully")
				}

				return nil
			}
		}

		return errors.New("service not found")
	},
}

func init() {
	nlbServiceDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	nlbServiceDeleteCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbServiceCmd.AddCommand(nlbServiceDeleteCmd)
}
