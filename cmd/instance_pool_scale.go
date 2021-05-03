package cmd

import (
	"errors"
	"fmt"
	"strconv"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var instancePoolScaleCmd = &cobra.Command{
	Use:   "scale NAME|ID SIZE",
	Short: "Scale an Instance Pool size",
	Long: `This command scales an Instance Pool size up (growing) or down
(shrinking).

In case of a scale-down, operators should use the "exo instancepool evict"
variant, allowing them to specify which specific instance should be evicted
from the Instance Pool rather than leaving the decision to the orchestrator.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		size, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid size %q", args[1])
		}
		if size <= 0 {
			return errors.New("minimum Instance Pool size is 1")
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		instancePool, err := lookupInstancePool(ctx, zone, args[0])
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf(
				"Are you sure you want to scale Instance Pool %q to %d?",
				instancePool.Name,
				size),
			) {
				return nil
			}
		}

		decorateAsyncOperation(fmt.Sprintf("Scaling Instance Pool %q...", instancePool.Name), func() {
			err = instancePool.Scale(ctx, int64(size))
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
	instancePoolScaleCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	instancePoolScaleCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	instancePoolCmd.AddCommand(instancePoolScaleCmd)
}
