package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
)

var storageCORSDeleteCmd = &cobra.Command{
	Use:     "delete sos://BUCKET",
	Aliases: []string{"del"},
	Short:   "Delete the CORS configuration of a bucket",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete bucket %s CORS configuration?",
				bucket)) {
				return nil
			}
		}

		storage, err := sos.NewStorageClient(
			GContext,
			sos.ClientOptZoneFromBucket(GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.DeleteBucketCORS(GContext, bucket); err != nil {
			return fmt.Errorf("unable to delete bucket CORS configuration: %w", err)
		}

		if !globalstate.Quiet {
			fmt.Println("CORS configuration deleted successfully")
		}

		return nil
	},
}

func init() {
	storageCORSDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	storageCORSCmd.AddCommand(storageCORSDeleteCmd)
}
