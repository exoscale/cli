package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
	"github.com/spf13/cobra"
)

var storageUploadCmd = &cobra.Command{
	Use:     "upload FILE... sos://BUCKET/[PREFIX/]",
	Aliases: []string{"put"},
	Short:   "Upload files to a bucket",
	Long: `This command uploads local files to a bucket.

Examples:

    # Upload files at the root of the bucket
    exo storage upload a b c sos://my-bucket

    # Upload files in a directory (trailing "/" in destination)
    exo storage upload index.html sos://my-bucket/public/

    # Upload a file under a different name
    exo storage upload a.txt sos://my-bucket/z.txt

    # Upload a directory recursively
    exo storage upload -r my-files/ sos://my-bucket
`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[len(args)-1] = strings.TrimPrefix(args[len(args)-1], sos.BucketPrefix)

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string

			sources = args[:len(args)-1]
			dst     = args[len(args)-1]
		)

		acl, err := cmd.Flags().GetString("acl")
		if err != nil {
			return err
		}
		if acl != "" && !utils.IsInList(sos.ObjectCannedACLToStrings(), acl) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl, strings.Join(sos.ObjectCannedACLToStrings(), ", "))
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		dstParts := strings.SplitN(dst, "/", 2)
		bucket = dstParts[0]
		if len(dstParts) > 1 {
			// Tricky case: if the user specifies "<bucket>/" as destination,
			// strings.SplitN()'s result slice contains an empty string as last
			// item: in this case we set the prefix as "/" to mean the root of
			// the bucket.
			if dstParts[len(dstParts)-1] == "" {
				prefix = "/"
			} else {
				prefix = dstParts[1]
			}
		} else {
			prefix = "/"
		}

		storage, err := sos.NewStorageClient(
			gContext,
			sos.ClientOptZoneFromBucket(gContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		return storage.UploadFiles(gContext, sources, &sos.StorageUploadConfig{
			Bucket:    bucket,
			Prefix:    prefix,
			ACL:       acl,
			Recursive: recursive,
			DryRun:    dryRun,
		})
	},
}

func init() {
	storageUploadCmd.Flags().String("acl", "",
		fmt.Sprintf("canned ACL to set on object (%s)", strings.Join(sos.ObjectCannedACLToStrings(), "|")))
	storageUploadCmd.Flags().BoolP("dry-run", "n", false,
		"simulate files upload, don't actually do it")
	storageUploadCmd.Flags().BoolP("recursive", "r", false,
		"upload directories recursively")
	storageCmd.AddCommand(storageUploadCmd)
}
