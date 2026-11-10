package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/collections"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
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

	PreRunE: func(c *cobra.Command, args []string) error {
		if len(args) != 2 {
			clicmd.CmdExitOnUsageError(c, "invalid arguments")
		}

		permittedOps := collections.NewSet(objOwnershipStatus, objOwnershipObjectWriter, objOwnershipBucketOwnerEnforced, objOwnershipBucketOwnerPreferred)
		if !permittedOps.Contains(args[objOwnershipOpArgIndex]) {
			clicmd.CmdExitOnUsageError(c, "invalid operation")
		}

		args[objOwnershipBucketArgIndex] = strings.TrimPrefix(args[objOwnershipBucketArgIndex], sos.BucketPrefix)

		clicmd.CmdSetZoneFlagFromDefault(c)

		return clicmd.CmdCheckRequiredFlags(c, []string{zoneFlagLong})
	},

	RunE: func(c *cobra.Command, args []string) error {
		ownershipCommand := args[objOwnershipOpArgIndex]
		bucket := args[objOwnershipBucketArgIndex]

		zone, err := c.Flags().GetString(zoneFlagLong)
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

		switch ownershipCommand {
		case objOwnershipStatus:
			return utils.PrintOutput(storage.GetBucketObjectOwnershipInfo(c.Context(), bucket))
		case objOwnershipObjectWriter:
			return storage.SetBucketObjectOwnership(c.Context(), bucket, sos.ObjectOwnershipObjectWriter)
		case objOwnershipBucketOwnerPreferred:
			return storage.SetBucketObjectOwnership(c.Context(), bucket, sos.ObjectOwnershipBucketOwnerPreferred)
		case objOwnershipBucketOwnerEnforced:
			return storage.SetBucketObjectOwnership(c.Context(), bucket, sos.ObjectOwnershipBucketOwnerEnforced)
		}

		return fmt.Errorf("invalid operation")
	},
}

var storageBucketObjectOwnershipCmdLongHelp = func() string {
	return "Manage the Object Ownership setting of a Storage Bucket"
}
