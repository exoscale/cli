package cmd

import (
	"fmt"
	"strings"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

type storageDownloadConfig struct {
	bucket      string
	prefix      string
	source      string
	destination string
	objects     []*s3types.Object
	recursive   bool
	overwrite   bool
	dryRun      bool
}

var storageDownloadCmd = &cobra.Command{
	Use:     "download sos://BUCKET/[OBJECT|PREFIX/] [DESTINATION]",
	Aliases: []string{"get"},
	Short:   "Download files from a bucket",
	Long: `This command downloads files from a bucket.

If no destination argument is provided, files will be stored into the current
directory.

Examples:

    # Download a single file
    exo storage download sos://my-bucket/file-a

    # Download a single file and rename it locally
    exo storage download sos://my-bucket/file-a file-z

    # Download a prefix recursively
    exo storage download -r sos://my-bucket/public/ /tmp/public/
`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		// Append implicit root prefix ("/") if only a bucket name is specified in the source
		if !strings.Contains(args[0], "/") {
			args[0] = args[0] + "/"
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string

			src = args[0]
			dst = "./"
		)

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(src, "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			prefix = parts[1]
			if prefix == "" {
				prefix = "/"
			}
		}

		if len(args) == 2 {
			dst = args[1]
		}

		if strings.HasSuffix(src, "/") && !recursive {
			return fmt.Errorf("%q is a directory, use flag `-r` to download recursively", src)
		}

		storage, err := newStorageClient(
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		objects := make([]*s3types.Object, 0)
		if err := storage.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
			objects = append(objects, o)
			return nil
		}); err != nil {
			return fmt.Errorf("error listing objects: %s", err)
		}

		return storage.downloadFiles(&storageDownloadConfig{
			bucket:      bucket,
			prefix:      prefix,
			source:      src,
			objects:     objects,
			destination: dst,
			recursive:   recursive,
			overwrite:   force,
			dryRun:      dryRun,
		})
	},
}

func init() {
	storageDownloadCmd.Flags().BoolP("force", "f", false,
		"overwrite existing destination files")
	storageDownloadCmd.Flags().BoolP("dry-run", "n", false,
		"simulate files download, don't actually do it")
	storageDownloadCmd.Flags().BoolP("recursive", "r", false,
		"download prefix recursively")
	storageCmd.AddCommand(storageDownloadCmd)
}
