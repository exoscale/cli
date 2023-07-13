package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/flags"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
)

func init() {
	storageShowCmd := &cobra.Command{
		Use:   "show sos://BUCKET/[OBJECT]",
		Short: "Show a bucket/object details",
		Long: fmt.Sprintf(`This command lists Storage buckets and objects.

Supported output template annotations:

	* When showing a bucket: %s
	* When showing an object: %s`,
			strings.Join(output.TemplateAnnotations(&sos.ShowBucketOutput{}), ", "),
			strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", ")),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				cmdExitOnUsageError(cmd, "invalid arguments")
			}

			args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

			// TODO add version flag
			versionID, err := cmd.Flags().GetString(flags.VersionID)
			if err != nil {
				return err
			}

			if versionID == "" {
				return nil
			}

			return flags.ValidateVersionID(versionID)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				bucket string
				key    string
			)

			parts := strings.SplitN(args[0], "/", 2)
			bucket = parts[0]
			if len(parts) > 1 {
				key = parts[1]
			}

			storage, err := sos.NewStorageClient(
				gContext,
				sos.ClientOptZoneFromBucket(gContext, bucket),
			)
			if err != nil {
				return fmt.Errorf("unable to initialize storage client: %w", err)
			}

			versionID, err := cmd.Flags().GetString(flags.VersionID)
			if err != nil {
				return err
			}

			if key == "" {
				if versionID != "" {
					fmt.Println("Warning: version id is ignored when showing a bucket")
				}

				return printOutput(storage.ShowBucket(gContext, bucket))
			}

			return printOutput(storage.ShowObject(gContext, bucket, key, versionID))
		},
	}
	storageShowCmd.Flags().String(flags.VersionID, "", flags.VersionIDUsage)
	storageCmd.AddCommand(storageShowCmd)
}
