package cmd

import (
	"errors"
	"fmt"
	"github.com/exoscale/cli/cmd/sos_sync"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultParallelSosSync = 10
)

func sosSyncRunE(cmd *cobra.Command, args []string) error {
	certsFile, err := cmd.Flags().GetString("certs-file")
	if err != nil {
		return err
	}

	removeDeleted, err := cmd.Flags().GetBool("remove-deleted")
	if err != nil {
		return err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	concurrency, err := cmd.Flags().GetUint16("concurrency")
	if err != nil {
		return err
	}
	if concurrency < 1 {
		return errors.New("concurrency cannot be less than 1")
	}

	if len(args) < 3 {
		return cmd.Usage()
	}
	targetBucket := args[0]

	sourceDirectory, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}

	targetPath := strings.TrimLeft(args[2], "/") + "/"

	sosClient, err := newSOSClient(certsFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(sourceDirectory); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Source dirctory does not exist: %s", sourceDirectory))
	}

	var bucketExists bool
	bucketExists, err = sosClient.BucketExists(targetBucket)
	if err != nil {
		return err
	}
	if !bucketExists {
		return errors.New(fmt.Sprintf("The target bucket does not exist: %s", targetBucket))
	}

	location, err := sosClient.GetBucketLocation(targetBucket)
	if err != nil {
		return err
	}

	if err := sosClient.setZone(location); err != nil {
		return err
	}

	fileStorage := sos_sync.NewLocalFileStorage(
		filepath.ToSlash(sourceDirectory),
		dryRun,
	)
	objectStorage := sos_sync.NewMinioObjectStorageOverlay(
		sosClient.Client,
		dryRun,
		targetBucket,
		strings.TrimLeft(targetPath, "/"),
	)
	ui := sos_sync.NewMbpUiFactory(false)
	sync := sos_sync.NewSyncEngine(
		ui,
		objectStorage,
		fileStorage,
		int(concurrency),
	)

	return sync.Synchronize(gContext, removeDeleted)
}

func init() {
	cmd := &cobra.Command{
		Use:   "sync <bucket name> <local path> <remote-path>",
		Short: "Sync a local folder with the object storage",
		RunE:  sosSyncRunE,
	}
	cmd.Flags().BoolP("remove-deleted", "d", false, "Delete remote files not present locally")
	cmd.Flags().BoolP("dry-run", "n", false, "Don't actually modify files")
	cmd.Flags().Uint16P("concurrency", "c", defaultParallelSosSync, "Parallel threads to use for upload")
	sosCmd.AddCommand(cmd)
}
