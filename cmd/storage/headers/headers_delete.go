package headers

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type deleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Target string `cli-arg:"#" cli-usage:"sos://BUCKET/(OBJECT|PREFIX/)"`

	Recursive          bool `cli-short:"r" cli-usage:"delete headers recursively (with object prefix only)"`
	CacheControl       bool `cli-usage:"delete the \"Cache-Control\" header"`
	ContentDisposition bool `cli-usage:"delete the \"Content-Disposition\" header"`
	ContentEncoding    bool `cli-usage:"delete the \"Content-Encoding\" header"`
	ContentLanguage    bool `cli-usage:"delete the \"Content-Language\" header"`
	ContentType        bool `cli-usage:"delete the \"Content-Type\" header"`
	Expires            bool `cli-usage:"delete the \"Expires\" header"`
}

func (c *deleteCmd) CmdAliases() []string { return []string{"del"} }
func (c *deleteCmd) CmdShort() string     { return "Delete HTTP headers from an object" }
func (c *deleteCmd) CmdLong() string {
	return fmt.Sprintf(`This command deletes response HTTP headers from objects.

Example:

    exo storage headers delete sos://my-bucket/data.json \
        --cache-control \
        --expires

Note: the "Content-Type" header cannot be removed, it is reset to its default
value "application/binary".

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", "))
}

func (c *deleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
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

	// A header flag set to false doesn't count, there would be nothing to delete.
	if !c.CacheControl && !c.ContentDisposition && !c.ContentEncoding &&
		!c.ContentLanguage && !c.ContentType && !c.Expires {
		clicmd.CmdExitOnUsageError(cmd, "no header flag specified")
	}

	return nil
}

func (c *deleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var headers []string

	parts := strings.SplitN(c.Target, "/", 2)
	bucket, prefix := parts[0], parts[1]

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	for _, h := range []struct {
		header string
		set    bool
	}{
		{sos.ObjectHeaderCacheControl, c.CacheControl},
		{sos.ObjectHeaderContentDisposition, c.ContentDisposition},
		{sos.ObjectHeaderContentEncoding, c.ContentEncoding},
		{sos.ObjectHeaderContentLanguage, c.ContentLanguage},
		{sos.ObjectHeaderContentType, c.ContentType},
		{sos.ObjectHeaderExpires, c.Expires},
	} {
		if h.set {
			headers = append(headers, h.header)
		}
	}

	if err := storage.DeleteObjectsHeaders(exocmd.GContext, bucket, prefix, headers, c.Recursive); err != nil {
		return fmt.Errorf("unable to add headers to object: %w", err)
	}

	if !globalstate.Quiet && !c.Recursive && !strings.HasSuffix(prefix, "/") {
		return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
	}

	if !globalstate.Quiet {
		fmt.Println("Headers deleted successfully")
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &deleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
