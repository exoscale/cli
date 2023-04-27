package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var storageHeaderAddCmd = &cobra.Command{
	Use:   "add sos://BUCKET/(OBJECT|PREFIX/)",
	Short: "Add HTTP headers to an object",
	Long: fmt.Sprintf(`This command adds response HTTP headers to objects.

Example:

    exo storage headers add sos://my-bucket/data.json \
        --content-type=application/json \
        --cache-control=no-store

Note: adding an already existing header will overwrite its value.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		if !strings.Contains(args[0], "/") {
			cmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		if headers := storageHeadersFromCmdFlags(cmd.Flags()); headers == nil {
			cmdExitOnUsageError(cmd, "no header flag specified")
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

		storage, err := sos.NewStorageClient(
			gContext,
			sos.ClientOptZoneFromBucket(gContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		headers := storageHeadersFromCmdFlags(cmd.Flags())
		if err := storage.UpdateObjectsHeaders(gContext, bucket, prefix, headers, recursive); err != nil {
			return fmt.Errorf("unable to add headers to object: %w", err)
		}

		if !globalstate.Quiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return printOutput(storage.ShowObject(gContext, bucket, prefix))
		}

		if !globalstate.Quiet {
			fmt.Println("Headers added successfully")
		}

		return nil
	},
}

func init() {
	storageHeaderAddCmd.Flags().BoolP("recursive", "r", false,
		"add headers recursively (with object prefix only)")
	storageHeaderAddCmd.Flags().String(strings.ToLower(sos.ObjectHeaderCacheControl), "",
		`value for "Cache-Control" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(sos.ObjectHeaderContentDisposition), "",
		`value for "Content-Disposition" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(sos.ObjectHeaderContentEncoding), "",
		`value for "Content-Encoding" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(sos.ObjectHeaderContentLanguage), "",
		`value for "Content-Language" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(sos.ObjectHeaderContentType), "",
		`value for "Content-Type" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(sos.ObjectHeaderExpires), "",
		`value for "Expires" header`)
	storageHeaderCmd.AddCommand(storageHeaderAddCmd)
}

// storageHeadersFromCmdFlags returns a non-nil map if at least
// one of the header-related command flags is set.
func storageHeadersFromCmdFlags(flags *pflag.FlagSet) map[string]*string {
	var headers map[string]*string

	flags.VisitAll(func(flag *pflag.Flag) {
		switch flag.Name {
		case strings.ToLower(sos.ObjectHeaderCacheControl):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[sos.ObjectHeaderCacheControl] = aws.String(v)
			}

		case strings.ToLower(sos.ObjectHeaderContentDisposition):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[sos.ObjectHeaderContentDisposition] = aws.String(v)
			}

		case strings.ToLower(sos.ObjectHeaderContentEncoding):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[sos.ObjectHeaderContentEncoding] = aws.String(v)
			}

		case strings.ToLower(sos.ObjectHeaderContentLanguage):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[sos.ObjectHeaderContentLanguage] = aws.String(v)
			}

		case strings.ToLower(sos.ObjectHeaderContentType):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[sos.ObjectHeaderContentType] = aws.String(v)
			}

		case strings.ToLower(sos.ObjectHeaderExpires):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[sos.ObjectHeaderExpires] = aws.String(v)
			}

		default:
			return
		}
	})

	return headers
}
