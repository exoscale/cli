package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var instancePoolEvictCmd = &cobra.Command{
	Use:   "evict INSTANCE-POOL-NAME|ID INSTANCE-NAME|ID...",
	Short: "Evict Instance Pool members",
	Long: `This command evicts specific members from an Instance Pool, effectively
scaling down the Instance Pool similar to the "exo instancepool scale" command.`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			i         = args[0]
			instances = args[1:]
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
			if !askQuestion(fmt.Sprintf("Are you sure you want to evict %v from Instance Pool %q?", instances, i)) {
				return nil
			}
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		instancePool, err := lookupInstancePool(ctx, zone, i)
		if err != nil {
			return err
		}

		members := make([]string, len(instances))
		for i, n := range instances {
			instance, err := getVirtualMachineByNameOrID(n)
			if err != nil {
				return fmt.Errorf("invalid Instance %q: %s", n, err)
			}
			members[i] = instance.ID.String()
		}

		decorateAsyncOperation(fmt.Sprintf("Evicting Instances from Instance Pool %q...", instancePool.Name), func() {
			err = instancePool.EvictMembers(ctx, members)
		})
		if err != nil {
			return err
		}

		if !gQuiet {
			return output(showInstancePool(zone, instancePool.ID))
		}

		return nil
	},
}

func init() {
	instancePoolEvictCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	instancePoolEvictCmd.Flags().StringP("zone", "z", "", "Instance Pool zone")
	instancePoolCmd.AddCommand(instancePoolEvictCmd)
}
