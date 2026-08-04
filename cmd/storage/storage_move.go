package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

var storageMoveCmd = &cobra.Command{
	Use:     "move sos://BUCKET/[OBJECT|PREFIX/] sos://BUCKET/[OBJECT|PREFIX/]",
	Aliases: []string{"mv"},
	Short:   "Move objects within a bucket or across buckets",
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

    exo storage move sos://my-bucket/file-a sos://my-bucket/folder/file-a

    exo storage move sos://my-bucket/file-a sos://other-bucket/file-a

    exo storage move -r sos://my-bucket/prefix/ sos://other-bucket/prefix/

    exo storage move -n sos://my-bucket/file-a sos://other-bucket/file-a
`,
	PreRunE: validateStorageTransferCommand,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStorageTransfer(cmd, args, true)
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

func validateStorageTransferCommand(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	recursive, err := cmd.Flags().GetBool("recursive")
	if err != nil {
		return err
	}

	return validateStorageTransferArgs(args, recursive)
}

func validateStorageTransferArgs(args []string, recursive bool) error {
	srcBucket, srcKey := parseBucketKey(args[0])
	dstBucket, dstKey := parseBucketKey(args[1])
	isPrefix := strings.HasSuffix(srcKey, "/") || recursive

	if srcBucket == "" {
		return fmt.Errorf("source must include a bucket name: %s", args[0])
	}
	if dstBucket == "" {
		return fmt.Errorf("destination must include a bucket name: %s", args[1])
	}
	if srcKey == "" && dstKey == "" {
		return fmt.Errorf("at least one of source/destination must include an object key or prefix")
	}
	if !isPrefix && srcKey != "" && dstKey == "" {
		return fmt.Errorf("destination must include an object key when source is a single object: %s", args[1])
	}
	if srcBucket == dstBucket && srcKey == dstKey {
		return fmt.Errorf("source and destination must differ")
	}
	if isPrefix && srcBucket == dstBucket && (strings.HasPrefix(dstKey, srcKey) || strings.HasPrefix(srcKey, dstKey)) {
		return fmt.Errorf("source and destination prefixes must not overlap")
	}

	return nil
}

func parseBucketKey(url string) (bucket, key string) {
	url = strings.TrimPrefix(url, sos.BucketPrefix)
	parts := strings.SplitN(url, "/", 2)
	bucket = parts[0]
	if len(parts) > 1 {
		key = parts[1]
	}
	return
}

func runStorageTransfer(cmd *cobra.Command, args []string, move bool) error {
	recursive, err := cmd.Flags().GetBool("recursive")
	if err != nil {
		return err
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}
	multipartConcurrency, err := cmd.Flags().GetInt("multipart-concurrency")
	if err != nil {
		return err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	srcBucket, srcKey := parseBucketKey(args[0])
	dstBucket, dstKey := parseBucketKey(args[1])

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, srcBucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	action := cmd.Name()
	pastTense := "copied"
	if move {
		pastTense = "moved"
	}
	isPrefix := strings.HasSuffix(srcKey, "/") || recursive

	if !force && !dryRun && isPrefix {
		if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf(
			"Are you sure you want to %s all objects from %s%s/%s to %s%s/%s?",
			action, sos.BucketPrefix, srcBucket, srcKey, sos.BucketPrefix, dstBucket, dstKey)) {
			return nil
		}
	}

	if dryRun {
		fmt.Println("[DRY-RUN]")
	}

	if !isPrefix {
		return runSingleObjectTransfer(storage, srcBucket, srcKey, dstBucket, dstKey, action, pastTense, multipartConcurrency, verbose, dryRun, move)
	}

	return runPrefixTransfer(storage, srcBucket, srcKey, dstBucket, dstKey, action, pastTense, multipartConcurrency, recursive, verbose, dryRun, move)
}

func runSingleObjectTransfer(storage *sos.Client, srcBucket, srcKey, dstBucket, dstKey, action, pastTense string, multipartConcurrency int, verbose, dryRun, move bool) error {
	if srcKey == "" {
		return fmt.Errorf("source must be an object key, not just a bucket: use --recursive for bucket prefixes")
	}

	if dryRun {
		fmt.Printf("%s %s%s/%s -> %s%s/%s\n", action, sos.BucketPrefix, srcBucket, srcKey, sos.BucketPrefix, dstBucket, dstKey)
		return nil
	}

	if err := transferObject(storage, srcBucket, srcKey, dstBucket, dstKey, multipartConcurrency, verbose, move); err != nil {
		return fmt.Errorf("%s failed: %w", action, err)
	}

	if verbose {
		showObj, err := storage.ShowObject(exocmd.GContext, dstBucket, dstKey)
		if err == nil {
			fmt.Printf("%s: %s -> %s (%d bytes, %s)\n", pastTense, srcKey, showObj.URL, showObj.Size, showObj.LastModified)
		}
	}

	return nil
}

func runPrefixTransfer(storage *sos.Client, srcBucket, srcKey, dstBucket, dstKey, action, pastTense string, multipartConcurrency int, recursive, verbose, dryRun, move bool) error {
	var transferred, failed int
	err := storage.ForEachObject(exocmd.GContext, srcBucket, srcKey, recursive, func(o *types.Object) error {
		if o.Key == nil {
			return nil
		}

		srcObjectKey := *o.Key
		srcObjectKeyTrimmed := strings.TrimPrefix(srcObjectKey, srcKey)
		dstObjectKey := dstKey + srcObjectKeyTrimmed

		if dryRun {
			fmt.Printf("%s %s%s/%s -> %s%s/%s\n", action, sos.BucketPrefix, srcBucket, srcObjectKey, sos.BucketPrefix, dstBucket, dstObjectKey)
			return nil
		}

		if err := transferObject(storage, srcBucket, srcObjectKey, dstBucket, dstObjectKey, multipartConcurrency, verbose, move); err != nil {
			fmt.Fprintf(os.Stderr, "%s failed for %s: %v\n", action, srcObjectKey, err)
			failed++
			return nil
		}

		transferred++
		if verbose && !globalstate.Quiet {
			fmt.Printf("%s: %s%s/%s -> %s%s/%s\n", pastTense, sos.BucketPrefix, srcBucket, srcObjectKey, sos.BucketPrefix, dstBucket, dstObjectKey)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s failed: %w", action, err)
	}

	if failed > 0 {
		return fmt.Errorf("%d object(s) failed to %s", failed, action)
	}

	if transferred == 0 && !dryRun && !globalstate.Quiet {
		fmt.Printf("no objects exist at %q\n", srcKey)
	}

	if verbose && !globalstate.Quiet && transferred > 0 {
		fmt.Printf("%s %d objects\n", pastTense, transferred)
	}

	return nil
}

func transferObject(storage *sos.Client, srcBucket, srcKey, dstBucket, dstKey string, multipartConcurrency int, verbose, move bool) error {
	if move {
		return storage.MoveObject(exocmd.GContext, srcBucket, srcKey, dstBucket, dstKey, multipartConcurrency, verbose)
	}

	return storage.CopyObjectTo(exocmd.GContext, srcBucket, srcKey, dstBucket, dstKey, multipartConcurrency, verbose)
}
