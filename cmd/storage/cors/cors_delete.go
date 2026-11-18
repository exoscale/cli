package cors

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type deleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	Force bool `cli-short:"f" cli-usage:"attempt to perform the operation without prompting for confirmation"`
}

func (c *deleteCmd) CmdAliases() []string { return []string{"del"} }
func (c *deleteCmd) CmdShort() string     { return "Delete the CORS configuration of a bucket" }
func (c *deleteCmd) CmdLong() string      { return "" }

func (c *deleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *deleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	if !c.Force {
		if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf("Are you sure you want to delete bucket %s CORS configuration?",
			bucket)) {
			return nil
		}
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	if err := storage.DeleteBucketCORS(exocmd.GContext, bucket); err != nil {
		return fmt.Errorf("unable to delete bucket CORS configuration: %w", err)
	}

	if !globalstate.Quiet {
		fmt.Println("CORS configuration deleted successfully")
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &deleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
