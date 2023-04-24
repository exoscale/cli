package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
)

var storageDeleteCmd = &cobra.Command{
	Use:     "delete sos://BUCKET/[OBJECT|PREFIX/]",
	Aliases: []string{"del", "rm"},
	Short:   "Delete objects",
	Long: `This command deletes objects stored in a bucket.

If you want to target objects under a "directory" prefix, suffix the path
argument with "/":

    exo storage delete sos://my-bucket/
    exo storage delete -r sos://my-bucket/some-directory/
`,

	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		if !strings.Contains(args[0], "/") {
			args[0] = args[0] + "/"
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

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
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
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete %s%s/%s?",
				storageBucketPrefix, bucket, prefix)) {
				return nil
			}
		}

		storage, err := sos.NewStorageClient(
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		deleted, err := storage.DeleteObjects(bucket, prefix, recursive)
		if err != nil {
			return fmt.Errorf("unable to delete objects: %w", err)
		}

		if verbose {
			for _, o := range deleted {
				fmt.Println(aws.ToString(o.Key))
			}
		}

		return nil
	},
}

func init() {
	storageDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	storageDeleteCmd.Flags().BoolP("recursive", "r", false, "delete objects recursively")
	storageDeleteCmd.Flags().BoolP("verbose", "v", false, "output deleted objects")
	storageCmd.AddCommand(storageDeleteCmd)
}
