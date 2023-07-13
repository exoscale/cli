package cmd

import (
	"fmt"
	"strings"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/flags"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/pkg/storage/sos/object"
)

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

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		// Append implicit root prefix ("/") if only a bucket name is specified in the source
		if !strings.Contains(args[0], "/") {
			args[0] += "/"
		}

		if err := flags.ValidateTimestampFlags(cmd); err != nil {
			return err
		}

		return flags.ValidateVersionFlags(cmd, false)
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

		storage, err := sos.NewStorageClient(
			gContext,
			sos.ClientOptZoneFromBucket(gContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		// TODO download versions
		filters, err := flags.TranslateTimeFilterFlagsToFilterFuncs(cmd)
		if err != nil {
			return err
		}

		/*
			modifyVersions, err := cmd.Flags().GetBool(flags.Versions)
			if err != nil {
				return err
			}

			versionFilters, err := flags.TranslateVersionFilterFlagsToFilterFuncs(cmd)
			if err != nil {
				return err
			}
		*/

		objects := make([]*s3types.Object, 0)
		if err := storage.ForEachObject(gContext, bucket, prefix, recursive, func(o object.ObjectInterface) error {
			objects = append(objects, o.GetObject())
			return nil
		}, filters); err != nil {
			return fmt.Errorf("error listing objects: %s", err)
		}

		return storage.DownloadFiles(gContext, &sos.DownloadConfig{
			Bucket:      bucket,
			Prefix:      prefix,
			Source:      src,
			Objects:     objects,
			Destination: dst,
			Recursive:   recursive,
			Overwrite:   force,
			DryRun:      dryRun,
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
	flags.AddVersionsFlags(storageDownloadCmd)
	flags.AddTimeFilterFlags(storageDownloadCmd)
	storageCmd.AddCommand(storageDownloadCmd)
}
