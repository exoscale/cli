package lifecycle

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type showCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	Zone string `cli-short:"z" cli-usage:"zone"`
}

func (c *showCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *showCmd) CmdShort() string     { return "Retrieve lifecycle configuration" }
func (c *showCmd) CmdLong() string      { return "Retrieve lifecycle configuration" }

func (c *showCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *showCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptWithZone(c.Zone),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	o, err := storage.GetBucketLifecycle(exocmd.GContext, bucket)
	if err != nil {
		return err
	}

	return c.OutputFunc(o, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &showCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
