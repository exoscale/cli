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

var storageMetadataDeleteCmd = &cobra.Command{
	Use:     "delete sos://BUCKET/(OBJECT|PREFIX/) KEY...",
	Aliases: []string{"del"},
	Short:   "Delete metadata from an object",
	Long: fmt.Sprintf(`This command deletes key/value metadata from an object.

Example:

    exo storage metadata delete sos://my-bucket/object-a k1

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		if !strings.Contains(args[0], "/") {
			exocmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket, prefix = parts[0], parts[1]
		mdKeys := args[1:]

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.DeleteObjectsMetadata(exocmd.GContext, bucket, prefix, mdKeys, recursive); err != nil {
			return fmt.Errorf("unable to delete metadata from object: %w", err)
		}

		if !globalstate.Quiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
		}

		if !globalstate.Quiet {
			fmt.Println("Metadata deleted successfully")
		}

		return nil
	},
}

func init() {
	storageMetadataDeleteCmd.Flags().BoolP("recursive", "r", false,
		"delete metadata recursively (with object prefix only)")
	storageMetadataCmd.AddCommand(storageMetadataDeleteCmd)
}
