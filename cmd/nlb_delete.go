package cmd

import (
	"fmt"

	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var nlbDeleteCmd = &cobra.Command{
	Use:     "delete <ID>",
	Short:   "Delete a Network Load Balancer",
	Aliases: gRemoveAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var nlbID = args[0]

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Do you really want to delete Network Load Balancer %q?", nlbID)) {
				return nil
			}
		}

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, ""))
		if err := cs.DeleteNetworkLoadBalancer(ctx, zone, nlbID); err != nil {
			return fmt.Errorf("unable to delete Network Load Balancer: %s", err)
		}

		if !gQuiet {
			cmd.Println("Network Load Balancer deleted successfully")
		}

		return nil
	},
}

func init() {
	nlbDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to delete without prompting for confirmation")
	nlbDeleteCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbCmd.AddCommand(nlbDeleteCmd)
}
