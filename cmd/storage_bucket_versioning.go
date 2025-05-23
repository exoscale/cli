package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/collections"
	"github.com/exoscale/cli/pkg/storage/sos"
)

const (
	objVersioningOpArgIndex     = 0
	objVersioningBucketArgIndex = 1
	objVersioningStatus         = "status"
	objVersioningEnable         = "enable"
	objVersioningSuspend        = "suspend"
)

func init() {
	storageBucketObjectVersioningCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketObjectVersioningCmd)
}

var storageBucketObjectVersioningCmd = &cobra.Command{
	Use:     "versioning {" + objVersioningStatus + "," + objVersioningEnable + "," + objVersioningSuspend + "} sos://BUCKET",
	Aliases: []string{"v"},
	Short:   "Manage the Object Versioning setting of a Storage Bucket",
	Long:    storageBucketObjectVersioningCmdLongHelp(),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		permittedOps := collections.NewSet(objVersioningStatus, objVersioningEnable, objVersioningSuspend)
		if !permittedOps.Contains(args[objOwnershipOpArgIndex]) {
			cmdExitOnUsageError(cmd, "invalid operation")
		}

		args[objVersioningBucketArgIndex] = strings.TrimPrefix(args[objVersioningBucketArgIndex], sos.BucketPrefix)

		CmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{zoneFlagLong})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		versioningCommand := args[objVersioningOpArgIndex]
		bucket := args[objVersioningBucketArgIndex]

		zone, err := cmd.Flags().GetString(zoneFlagLong)
		if err != nil {
			return err
		}

		storage, err := sos.NewStorageClient(
			GContext,
			sos.ClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		switch versioningCommand {
		case objVersioningStatus:
			return printOutput(storage.BucketVersioningStatus(cmd.Context(), bucket))
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
