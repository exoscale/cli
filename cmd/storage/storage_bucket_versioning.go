package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/collections"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

const (
	objVersioningOpArgIndex     = 0
	objVersioningBucketArgIndex = 1
	objVersioningStatus         = "status"
	objVersioningEnable         = "enable"
	objVersioningSuspend        = "suspend"
)

func init() {
	storageBucketObjectVersioningCmd.Flags().StringP(exocmd.ZoneFlagLong, exocmd.ZoneFlagShort, "", exocmd.ZoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketObjectVersioningCmd)
}

var storageBucketObjectVersioningCmd = &cobra.Command{
	Use:     "versioning {" + objVersioningStatus + "," + objVersioningEnable + "," + objVersioningSuspend + "} sos://BUCKET",
	Aliases: []string{"v"},
	Short:   "Manage the Object Versioning setting of a Storage Bucket",
	Long:    storageBucketObjectVersioningCmdLongHelp(),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		permittedOps := collections.NewSet(objVersioningStatus, objVersioningEnable, objVersioningSuspend)
		if !permittedOps.Contains(args[objOwnershipOpArgIndex]) {
			exocmd.CmdExitOnUsageError(cmd, "invalid operation")
		}

		args[objVersioningBucketArgIndex] = strings.TrimPrefix(args[objVersioningBucketArgIndex], sos.BucketPrefix)

		exocmd.CmdSetZoneFlagFromDefault(cmd)

		return exocmd.CmdCheckRequiredFlags(cmd, []string{exocmd.ZoneFlagLong})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		versioningCommand := args[objVersioningOpArgIndex]
		bucket := args[objVersioningBucketArgIndex]

		zone, err := cmd.Flags().GetString(exocmd.ZoneFlagLong)
		if err != nil {
			return err
		}

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		switch versioningCommand {
		case objVersioningStatus:
			return utils.PrintOutput(storage.BucketVersioningStatus(cmd.Context(), bucket))
		case objVersioningEnable:
			return storage.EnableBucketVersioning(cmd.Context(), bucket)
		case objVersioningSuspend:
			return storage.SuspendBucketVersioning(cmd.Context(), bucket)
		}

		return fmt.Errorf("invalid operation")
	},
}

var storageBucketObjectVersioningCmdLongHelp = func() string {
	return "Manage the Object Versioning setting of a Storage Bucket"
}
