package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
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
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		if !strings.Contains(args[0], "/") {
			exocmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		for _, kv := range args[1:] {
			if !strings.Contains(kv, "=") {
				exocmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", kv))
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
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.AddObjectsMetadata(exocmd.GContext, bucket, prefix, metadata, recursive); err != nil {
			return fmt.Errorf("unable to add metadata to object: %w", err)
		}

		if !globalstate.Quiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
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
