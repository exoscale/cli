package headers

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type addCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	Target string `cli-arg:"#" cli-usage:"sos://BUCKET/(OBJECT|PREFIX/)"`

	Recursive          bool   `cli-short:"r" cli-usage:"add headers recursively (with object prefix only)"`
	CacheControl       string `cli-usage:"value for \"Cache-Control\" header"`
	ContentDisposition string `cli-usage:"value for \"Content-Disposition\" header"`
	ContentEncoding    string `cli-usage:"value for \"Content-Encoding\" header"`
	ContentLanguage    string `cli-usage:"value for \"Content-Language\" header"`
	ContentType        string `cli-usage:"value for \"Content-Type\" header"`
	Expires            string `cli-usage:"value for \"Expires\" header"`
}

func (c *addCmd) CmdAliases() []string { return nil }
func (c *addCmd) CmdShort() string     { return "Add HTTP headers to an object" }
func (c *addCmd) CmdLong() string {
	return fmt.Sprintf(`This command adds response HTTP headers to objects.

Example:

    exo storage headers add sos://my-bucket/data.json \
        --content-type=application/json \
        --cache-control=no-store

Note: adding an already existing header will overwrite its value.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", "))
}

func (c *addCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

	if !strings.Contains(args[0], "/") {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	// A header flag only counts when its value is not empty.
	if c.CacheControl == "" && c.ContentDisposition == "" && c.ContentEncoding == "" &&
		c.ContentLanguage == "" && c.ContentType == "" && c.Expires == "" {
		clicmd.CmdExitOnUsageError(cmd, "no header flag specified")
	}

	return nil
}

func (c *addCmd) CmdRun(_ *cobra.Command, _ []string) error {
	parts := strings.SplitN(c.Target, "/", 2)
	bucket, prefix := parts[0], parts[1]

	headers := make(map[string]*string)
	for header, value := range map[string]string{
		sos.ObjectHeaderCacheControl:       c.CacheControl,
		sos.ObjectHeaderContentDisposition: c.ContentDisposition,
		sos.ObjectHeaderContentEncoding:    c.ContentEncoding,
		sos.ObjectHeaderContentLanguage:    c.ContentLanguage,
		sos.ObjectHeaderContentType:        c.ContentType,
		sos.ObjectHeaderExpires:            c.Expires,
	} {
		if value != "" {
			headers[header] = aws.String(value)
		}
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	if err := storage.UpdateObjectsHeaders(exocmd.GContext, bucket, prefix, headers, c.Recursive); err != nil {
		return fmt.Errorf("unable to add headers to object: %w", err)
	}

	if !globalstate.Quiet && !c.Recursive && !strings.HasSuffix(prefix, "/") {
		return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
	}

	if !globalstate.Quiet {
		fmt.Println("Headers added successfully")
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
