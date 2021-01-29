package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var instancePoolDeleteCmd = &cobra.Command{
	Use:     "delete <name | id>+",
	Short:   "Delete an instance pool",
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

		zoneName, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		zone, err := getZoneByNameOrID(zoneName)
		if err != nil {
			return err
		}

		// Ensure the Instance Pool is not attached to a NLB service.
		nlbs, err := cs.ListNetworkLoadBalancers(gContext, zone.Name)
		if err != nil {
			return fmt.Errorf("unable to list Network Load Balancers: %v", err)
		}

		tasks := make([]task, 0, len(args))
		for _, arg := range args {
			if !force {
				if !askQuestion(fmt.Sprintf("sure you want to delete %q", arg)) {
					continue
				}
			}

			i, err := getInstancePoolByNameOrID(arg, zone.ID)
			if err != nil {
				return err
			}

			for _, nlb := range nlbs {
				for _, svc := range nlb.Services {
					if svc.InstancePoolID == i.ID.String() {
						return fmt.Errorf("instance pool %q is still referenced by NLB service %s/%s",
							i.Name, nlb.Name, svc.Name)
					}
				}
			}

			tasks = append(tasks, task{
				egoscale.DestroyInstancePool{
					ID:     i.ID,
					ZoneID: zone.ID,
				},
				fmt.Sprintf("Deleting instance pool %q", args[0]),
			})
		}

		r := asyncTasks(tasks)
		errs := filterErrors(r)
		if len(errs) > 0 {
			return errs[0]
		}

		return nil
	},
}

func init() {
	instancePoolDeleteCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove instance pool without prompting for confirmation")
	instancePoolCmd.AddCommand(instancePoolDeleteCmd)
}
