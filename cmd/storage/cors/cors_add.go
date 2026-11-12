package cors

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type addCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	AllowedOrigin []string `cli-usage:"allowed origin (can be repeated multiple times)"`
	AllowedMethod []string `cli-usage:"allowed method (can be repeated multiple times)"`
	AllowedHeader []string `cli-usage:"allowed header (can be repeated multiple times)"`
}

func (c *addCmd) CmdAliases() []string { return nil }
func (c *addCmd) CmdShort() string     { return "Add a CORS configuration rule to a bucket" }
func (c *addCmd) CmdLong() string {
	return `This command adds a new rule to the current bucket CORS
configuration.

Example:

    exo storage cors add sos://my-bucket \
        --allowed-origin "https://my-website.net" \
        --allowed-method "*"
`
}

func (c *addCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"allowed-origin", "allowed-method"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	rule := &sos.CORSRule{
		AllowedOrigins: c.AllowedOrigin,
		AllowedMethods: c.AllowedMethod,
		AllowedHeaders: c.AllowedHeader,
	}
	if err := storage.AddBucketCORSRule(exocmd.GContext, bucket, rule); err != nil {
		return fmt.Errorf("unable to add rule to the bucket CORS configuration: %w", err)
	}

	if !globalstate.Quiet {
		return c.OutputFunc(storage.ShowBucket(exocmd.GContext, bucket))
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
