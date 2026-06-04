package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

var storageMoveCmd = &cobra.Command{
	Use:   "move sos://BUCKET/[OBJECT|PREFIX/] sos://BUCKET/[OBJECT|PREFIX/]",
	Short: "Move objects within a bucket or across buckets",
	Long: `Move objects within a bucket or across buckets.

This command moves objects by performing a server-side copy followed by
a delete of the source object. Object metadata, headers, and ACLs are
preserved.

Warning: move is implemented as server-side copy followed by delete.
If the delete step fails after a successful copy, the object will
remain in both locations. There is no automatic rollback.

Multi-object prefix moves are processed serially. A trailing slash on the
source selects prefix mode; -r controls recursion into subdirectories.

Examples:

    exo storage move sos://my-bucket/file-a sos://my-bucket/folder/

    exo storage move sos://my-bucket/file-a sos://other-bucket/file-a

    exo storage move -r sos://my-bucket/prefix/ sos://other-bucket/prefix/

    exo storage move -n sos://my-bucket/file-a sos://other-bucket/
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return cmd.Usage()
		}
		return validateMoveArgs(args)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		srcBucket, srcKey := parseBucketKey(args[0])
		dstBucket, dstKey := parseBucketKey(args[1])

		recursive, _ := cmd.Flags().GetBool("recursive")
		force, _ := cmd.Flags().GetBool("force")
		multipartConcurrency, _ := cmd.Flags().GetInt("multipart-concurrency")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, srcBucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		isPrefix := strings.HasSuffix(srcKey, "/") || recursive

		if !force && !dryRun && isPrefix {
			if !confirmPrefixMove(exocmd.GContext, srcBucket, srcKey, dstBucket, dstKey) {
				return nil
			}
		}

		if dryRun {
			fmt.Println("[DRY-RUN]")
		}

		if !isPrefix {
			return runSingleObjectMove(storage, srcBucket, srcKey, dstBucket, dstKey, multipartConcurrency, verbose, dryRun)
		}

		return runPrefixMove(storage, srcBucket, srcKey, dstBucket, dstKey, multipartConcurrency, recursive, verbose, dryRun)
	},
}

func init() {
	storageMoveCmd.Flags().BoolP("dry-run", "n", false, "simulate the move operation")
	storageMoveCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
	storageMoveCmd.Flags().BoolP("recursive", "r", false, "move objects recursively")
	storageMoveCmd.Flags().BoolP("verbose", "v", false, "output moved objects")
	storageMoveCmd.Flags().Int("multipart-concurrency", 4, "number of concurrent parts for multipart moves")
	storageCmd.AddCommand(storageMoveCmd)
}
