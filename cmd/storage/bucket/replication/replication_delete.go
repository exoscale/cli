package replication

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type deleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	Zone string `cli-short:"z" cli-usage:"bucket zone"`
}

func (c *deleteCmd) CmdAliases() []string { return nil }
func (c *deleteCmd) CmdShort() string     { return "Delete replication configuration" }
func (c *deleteCmd) CmdLong() string      { return "" }

func (c *deleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if err := checkArgsAndZone(cmd, args, 1); err != nil {
		return err
	}
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *deleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptWithZone(c.Zone),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	return storage.DeleteBucketReplication(exocmd.GContext, bucket)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &deleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
