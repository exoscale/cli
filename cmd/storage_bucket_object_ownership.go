package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/collections"
	"github.com/exoscale/cli/pkg/storage/sos"
)

const (
	objOwnershipOpArgIndex           = 0
	objOwnershipBucketArgIndex       = 1
	objOwnershipStatus               = "status"
	objOwnershipObjectWriter         = "object-writer"
	objOwnershipBucketOwnerEnforced  = "bucket-owner-enforced"
	objOwnershipBucketOwnerPreferred = "bucket-owner-preferred"
)

func init() {
	storageBucketObjectOwnershipCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketObjectOwnershipCmd)
}

var storageBucketObjectOwnershipCmd = &cobra.Command{
	Use:     "object-ownership {" + objOwnershipStatus + "," + objOwnershipObjectWriter + "," + objOwnershipBucketOwnerEnforced + "," + objOwnershipBucketOwnerPreferred + "} sos://BUCKET",
	Aliases: []string{"oo"},
	Short:   "Manage the Object Ownership setting of a Storage Bucket",
	Long:    storageBucketObjectOwnershipCmdLongHelp(),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		permittedOps := collections.NewSet(objOwnershipStatus, objOwnershipObjectWriter, objOwnershipBucketOwnerEnforced, objOwnershipBucketOwnerPreferred)
		if !permittedOps.Contains(args[objOwnershipOpArgIndex]) {
			cmdExitOnUsageError(cmd, "invalid operation")
		}

		args[objOwnershipBucketArgIndex] = strings.TrimPrefix(args[objOwnershipBucketArgIndex], sos.BucketPrefix)

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{zoneFlagLong})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		ownershipCommand := args[objOwnershipOpArgIndex]
		bucket := args[objOwnershipBucketArgIndex]

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

		switch ownershipCommand {
		case objOwnershipStatus:
			return printOutput(storage.GetBucketObjectOwnershipInfo(cmd.Context(), bucket))
		case objOwnershipObjectWriter:
			return storage.SetBucketObjectOwnership(cmd.Context(), bucket, sos.ObjectOwnershipObjectWriter)
		case objOwnershipBucketOwnerPreferred:
			return storage.SetBucketObjectOwnership(cmd.Context(), bucket, sos.ObjectOwnershipBucketOwnerPreferred)
		case objOwnershipBucketOwnerEnforced:
			return storage.SetBucketObjectOwnership(cmd.Context(), bucket, sos.ObjectOwnershipBucketOwnerEnforced)
		}

		return fmt.Errorf("invalid operation")
	},
}

var storageBucketObjectOwnershipCmdLongHelp = func() string {
	return "Manage the Object Ownership setting of a Storage Bucket"
}
