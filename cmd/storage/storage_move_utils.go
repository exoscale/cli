package storage

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

func validateMoveArgs(args []string) error {
	srcBucket, srcKey := parseBucketKey(args[0])
	dstBucket, dstKey := parseBucketKey(args[1])

	if srcBucket == "" {
		return fmt.Errorf("source must include a bucket name: %s", args[0])
	}
	if dstBucket == "" {
		return fmt.Errorf("destination must include a bucket name: %s", args[1])
	}
	if srcKey == "" && dstKey == "" {
		return fmt.Errorf("at least one of source/destination must include an object key or prefix")
	}
	if srcKey != "" && dstKey == "" {
		return fmt.Errorf("destination must include an object key when source is a single object: %s", args[1])
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

func confirmPrefixMove(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) bool {
	return utils.AskQuestion(ctx, fmt.Sprintf(
		"Are you sure you want to move all objects from %s%s/%s to %s%s/%s?",
		sos.BucketPrefix, srcBucket, srcKey, sos.BucketPrefix, dstBucket, dstKey))
}

func runSingleObjectMove(storage *sos.Client, srcBucket, srcKey, dstBucket, dstKey string, multipartConcurrency int, verbose, dryRun bool) error {
	if srcKey == "" {
		return fmt.Errorf("source must be an object key, not just a bucket: use a trailing slash for prefix moves")
	}

	if dryRun {
		fmt.Printf("move %s%s/%s -> %s%s/%s\n", sos.BucketPrefix, srcBucket, srcKey, sos.BucketPrefix, dstBucket, dstKey)
		return nil
	}

	if err := storage.MoveObject(exocmd.GContext, srcBucket, srcKey, dstBucket, dstKey, multipartConcurrency, verbose); err != nil {
		return fmt.Errorf("move failed: %w", err)
	}

	if verbose {
		showObj, err := storage.ShowObject(exocmd.GContext, dstBucket, dstKey)
		if err == nil {
			fmt.Printf("moved: %s -> %s (%d bytes, %s)\n", srcKey, showObj.URL, showObj.Size, showObj.LastModified)
		}
	}

	return nil
}

func runPrefixMove(storage *sos.Client, srcBucket, srcKey, dstBucket, dstKey string, multipartConcurrency int, recursive, verbose, dryRun bool) error {
	var moved, failed int
	err := storage.ForEachObject(exocmd.GContext, srcBucket, srcKey, recursive, func(o *types.Object) error {
		if o.Key == nil {
			return nil
		}

		srcObjectKey := *o.Key
		srcObjectKeyTrimmed := strings.TrimPrefix(srcObjectKey, srcKey)
		dstObjectKey := dstKey + srcObjectKeyTrimmed

		if dryRun {
			fmt.Printf("move %s%s/%s -> %s%s/%s\n", sos.BucketPrefix, srcBucket, srcObjectKey, sos.BucketPrefix, dstBucket, dstObjectKey)
			return nil
		}

		if err := storage.MoveObject(exocmd.GContext, srcBucket, srcObjectKey, dstBucket, dstObjectKey, multipartConcurrency, verbose); err != nil {
			fmt.Fprintf(os.Stderr, "move failed for %s: %v\n", srcObjectKey, err)
			failed++
			return nil
		}

		moved++
		if verbose && !globalstate.Quiet {
			fmt.Printf("moved: %s%s/%s -> %s%s/%s\n", sos.BucketPrefix, srcBucket, srcObjectKey, sos.BucketPrefix, dstBucket, dstObjectKey)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("move failed: %w", err)
	}

	if failed > 0 {
		return fmt.Errorf("%d object(s) failed to move", failed)
	}

	if moved == 0 && !dryRun && !globalstate.Quiet {
		fmt.Printf("no objects exist at %q\n", srcKey)
	}

	if verbose && !globalstate.Quiet && moved > 0 {
		fmt.Printf("moved %d objects\n", moved)
	}

	return nil
}
