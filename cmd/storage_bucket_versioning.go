package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

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

type empty struct{}

type Set[T comparable] map[T]empty

func storeValuesInMap[T comparable](values ...T) Set[T] {
	result := make(Set[T])
	for _, value := range values {
		result[value] = empty{}
	}

	return result
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

		permittedOps := storeValuesInMap(objVersioningStatus, objVersioningEnable, objVersioningSuspend)
		if _, ok := permittedOps[args[objVersioningOpArgIndex]]; !ok {
			cmdExitOnUsageError(cmd, "invalid operation")
		}

		args[objVersioningBucketArgIndex] = strings.TrimPrefix(args[objVersioningBucketArgIndex], sos.BucketPrefix)

		cmdSetZoneFlagFromDefault(cmd)

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
			gContext,
			sos.ClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		switch versioningCommand {
		case objVersioningStatus:
			return printOutput(storage.GetBucketVersioningSetting(cmd.Context(), bucket))
		case objVersioningEnable:
			return storage.EnableBucketVersioningSetting(cmd.Context(), bucket)
		case objVersioningSuspend:
			return storage.SuspendBucketVersioningSetting(cmd.Context(), bucket)
		}

		return fmt.Errorf("invalid operation")
	},
}

var storageBucketObjectVersioningCmdLongHelp = func() string {
	return "Manage the Object Versioning setting of a Storage Bucket"
}
