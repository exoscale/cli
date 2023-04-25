package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
)

var storageMetadataAddCmd = &cobra.Command{
	Use:   "add sos://BUCKET/(OBJECT|PREFIX/) KEY=VALUE...",
	Short: "Add key/value metadata to an object",
	Long: fmt.Sprintf(`This command adds key/value metadata to an object.

Example:

    exo storage metadata add sos://my-bucket/object-a \
        k1=v1 \
        k2=v2

Notes:

  * Adding an already existing key will overwrite its value.
  * The following characters are not allowed in keys: %s

Supported output template annotations: %s`,
		strings.Join(strings.Split(sos.MetadataForbiddenCharset, ""), " "),
		strings.Join(output.OutputterTemplateAnnotations(&sos.ShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		if !strings.Contains(args[0], "/") {
			cmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		for _, kv := range args[1:] {
			if !strings.Contains(kv, "=") {
				cmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", kv))
			}
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket   string
			prefix   string
			metadata = make(map[string]string)
		)

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket, prefix = parts[0], parts[1]

		for _, kv := range args[1:] {
			parts = strings.Split(kv, "=")
			metadata[parts[0]] = parts[1]
		}

		storage, err := sos.NewStorageClient(
			gContext,
			sos.ClientOptZoneFromBucket(gContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.AddObjectsMetadata(gContext, bucket, prefix, metadata, recursive); err != nil {
			return fmt.Errorf("unable to add metadata to object: %w", err)
		}

		if !globalstate.Quiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return printOutput(storage.ShowObject(gContext, bucket, prefix))
		}

		if !globalstate.Quiet {
			fmt.Println("Metadata added successfully")
		}

		return nil
	},
}

func init() {
	storageMetadataAddCmd.Flags().BoolP("recursive", "r", false,
		"add metadata recursively (with object prefix only)")
	storageMetadataCmd.AddCommand(storageMetadataAddCmd)
}
