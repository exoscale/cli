package replication

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type showCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	Zone string `cli-short:"z" cli-usage:"bucket zone"`
}

func (c *showCmd) CmdAliases() []string { return nil }
func (c *showCmd) CmdShort() string     { return "Retrieve replication configuration" }
func (c *showCmd) CmdLong() string      { return "" }

func (c *showCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if err := checkArgsAndZone(cmd, args, 1); err != nil {
		return err
	}
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
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

	o, err := storage.GetBucketReplication(exocmd.GContext, bucket)
	if err != nil {
		return err
	}

	return c.OutputFunc(o, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &showCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
