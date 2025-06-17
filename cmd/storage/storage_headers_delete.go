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

var storageHeaderDeleteCmd = &cobra.Command{
	Use:     "delete sos://BUCKET/(OBJECT|PREFIX/)",
	Aliases: []string{"del"},
	Short:   "Delete HTTP headers from an object",
	Long: fmt.Sprintf(`This command deletes response HTTP headers from objects.

Example:

    exo storage headers delete sos://my-bucket/data.json \
        --cache-control \
        --expires

Note: the "Content-Type" header cannot be removed, it is reset to its default
value "application/binary".

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		if !strings.Contains(args[0], "/") {
			exocmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		var hasHeaderFlagsSet bool
		for _, flag := range []string{
			sos.ObjectHeaderCacheControl,
			sos.ObjectHeaderContentDisposition,
			sos.ObjectHeaderContentEncoding,
			sos.ObjectHeaderContentLanguage,
			sos.ObjectHeaderContentType,
			sos.ObjectHeaderExpires,
		} {
			if cmd.Flags().Changed(strings.ToLower(flag)) {
				hasHeaderFlagsSet = true
				break
			}
		}
		if !hasHeaderFlagsSet {
			exocmd.CmdExitOnUsageError(cmd, "no header flag specified")
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket  string
			prefix  string
			headers []string
		)

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket, prefix = parts[0], parts[1]

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		for _, header := range []string{
			sos.ObjectHeaderCacheControl,
			sos.ObjectHeaderContentDisposition,
			sos.ObjectHeaderContentEncoding,
			sos.ObjectHeaderContentLanguage,
			sos.ObjectHeaderContentType,
			sos.ObjectHeaderExpires,
		} {
			if ok, _ := cmd.Flags().GetBool(strings.ToLower(header)); ok {
				headers = append(headers, header)
			}
		}

		if err := storage.DeleteObjectsHeaders(exocmd.GContext, bucket, prefix, headers, recursive); err != nil {
			return fmt.Errorf("unable to add headers to object: %w", err)
		}

		if !globalstate.Quiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
		}

		if !globalstate.Quiet {
			fmt.Println("Headers deleted successfully")
		}

		return nil
	},
}

func init() {
	storageHeaderDeleteCmd.Flags().BoolP("recursive", "r", false,
		"delete headers recursively (with object prefix only)")
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(sos.ObjectHeaderCacheControl), false,
		`delete the "Cache-Control" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(sos.ObjectHeaderContentDisposition), false,
		`delete the "Content-Disposition" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(sos.ObjectHeaderContentEncoding), false,
		`delete the "Content-Encoding" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(sos.ObjectHeaderContentLanguage), false,
		`delete the "Content-Language" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(sos.ObjectHeaderContentType), false,
		`delete the "Content-Type" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(sos.ObjectHeaderExpires), false,
		`delete the "Expires" header`)
	storageHeaderCmd.AddCommand(storageHeaderDeleteCmd)
}
