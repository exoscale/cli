package storage

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

var storagePurgeCmd = &cobra.Command{
	Use:     "purge sos://BUCKET/[OBJECT|PREFIX/]",
	Aliases: []string{"purge"},
	Short:   "Purge objects and their versions",
	Long: `This command purges objects and object versions stored in a bucket.

    exo storage purge sos://my-bucket/
    exo storage purge sos://my-bucket/some-prefix/
`,

	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		if !strings.Contains(args[0], "/") {
			args[0] += "/"
		}
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		parts := strings.SplitN(args[0], "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			prefix = parts[1]

			// Special case: the caller wants to target objects at the root of
			// the bucket, in this case the prefix is empty so we set the key
			// to a symbolic value that shall be removed later on.
			if prefix == "" {
				prefix = "/"
			}
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf("Are you sure you want to delete %s%s/%s?",
				sos.BucketPrefix, bucket, prefix)) {
				return nil
			}
		}

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		deletedChan, errChan := storage.DeleteObjectVersions(exocmd.GContext, bucket, prefix)

		for {
			select {
			case err, ok := <-errChan:
				if ok {
					fmt.Printf("Error happened: %v\n", err)
				} else {
					fmt.Println("Purge completed")
					return nil
				}
			case deletedElt := <-deletedChan:
				if verbose {
					fmt.Println("deleted:", aws.ToString(deletedElt.Key))
				}
			}
		}
	},
}

func init() {
	storagePurgeCmd.Flags().BoolP("force", "f", false, exocmd.CmdFlagForceHelp)
	storagePurgeCmd.Flags().BoolP("verbose", "v", false, "output deleted objects")
	storageCmd.AddCommand(storagePurgeCmd)
}
