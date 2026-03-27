package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type storageMoveLocation struct {
	Bucket string
	Key    string
}

var storageMoveCmd = &cobra.Command{
	Use:     "move SOURCE DESTINATION",
	Aliases: []string{"mv"},
	Short:   "Move objects without local download",
	Long: `This command moves objects between buckets without downloading them locally.

Examples:

    # Move an object within a bucket
    exo storage move sos://my-bucket/file-a sos://my-bucket/archive/file-a

    # Move an object to another bucket and keep its basename
    exo storage move sos://my-bucket/file-a sos://other-bucket/archive/

    # Move a prefix recursively
    exo storage move -r sos://my-bucket/public/ sos://other-bucket/archive/

Notes:

    * If the destination ends with "/", the source basename is preserved.
    * Prefix moves preserve relative paths under the destination prefix.
    * Existing destination objects are overwritten.
`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
		args[1] = strings.TrimPrefix(args[1], sos.BucketPrefix)

		if !strings.Contains(args[0], "/") {
			exocmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		source, err := parseStorageMoveLocation(args[0])
		if err != nil {
			return err
		}

		destination, err := parseStorageMoveLocation(args[1])
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		concurrency, err := cmd.Flags().GetInt("concurrency")
		if err != nil {
			return err
		}
		if concurrency < 1 {
			return fmt.Errorf("invalid concurrency %d, value must be greater than 0", concurrency)
		}

		srcStorage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, source.Bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize source storage client: %w", err)
		}

		dstStorage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, destination.Bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize destination storage client: %w", err)
		}

		moved, moveErr := dstStorage.MoveObjects(
			exocmd.GContext,
			srcStorage,
			source.Bucket,
			source.Key,
			destination.Bucket,
			destination.Key,
			&sos.StorageMoveConfig{
				Recursive:                recursive,
				DryRun:                   dryRun,
				MultipartCopyConcurrency: concurrency,
			},
		)

		if dryRun {
			fmt.Println("[DRY-RUN]")
		}

		if dryRun || verbose {
			for _, move := range moved {
				fmt.Printf("%s%s/%s -> %s%s/%s\n",
					sos.BucketPrefix, move.SourceBucket, move.SourceKey,
					sos.BucketPrefix, move.DestinationBucket, move.DestinationKey,
				)
			}
		}

		if moveErr != nil {
			if !dryRun && !globalstate.Quiet && len(moved) > 0 && !verbose {
				fmt.Printf("Moved %d object(s) before an error occurred\n", len(moved))
			}
			return moveErr
		}

		if len(moved) == 0 {
			if !globalstate.Quiet {
				fmt.Printf("no objects exist at %q\n", source.Key)
			}
			return nil
		}

		if dryRun {
			return nil
		}

		return nil
	},
}

func init() {
	storageMoveCmd.Flags().BoolP("dry-run", "n", false,
		"simulate object moves, don't actually do them")
	storageMoveCmd.Flags().BoolP("recursive", "r", false,
		"move object prefixes recursively")
	storageMoveCmd.Flags().IntP("concurrency", "c", 5,
		"number of parallel multipart copy workers for large object moves")
	storageMoveCmd.Flags().BoolP("verbose", "v", false,
		"output moved objects")
	storageCmd.AddCommand(storageMoveCmd)
}

func parseStorageMoveLocation(value string) (storageMoveLocation, error) {
	var location storageMoveLocation

	parts := strings.SplitN(value, "/", 2)
	location.Bucket = parts[0]
	if location.Bucket == "" {
		return location, fmt.Errorf("invalid bucket name")
	}

	if len(parts) == 1 {
		return location, nil
	}

	location.Key = parts[1]
	if location.Key == "" {
		location.Key = "/"
	}

	return location, nil
}
