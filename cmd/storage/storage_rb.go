package storage

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

type storageRbCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"rb"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	Force     bool `cli-short:"f" cli-usage:"attempt to perform the operation without prompting for confirmation"`
	Recursive bool `cli-short:"r" cli-usage:"empty the bucket before deleting it"`
}

func (c *storageRbCmd) CmdAliases() []string { return nil }
func (c *storageRbCmd) CmdShort() string     { return "Delete a bucket" }
func (c *storageRbCmd) CmdLong() string      { return "" }

func (c *storageRbCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	target := args[0]
	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
	args[0] = strings.TrimRight(args[0], "/")

	if args[0] == "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", target))
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *storageRbCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if !c.Force {
		if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf("Are you sure you want to delete %s%s?", sos.BucketPrefix, c.Bucket)) {
			return nil
		}
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, c.Bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	if err := storage.DeleteBucket(exocmd.GContext, c.Bucket, c.Recursive); err != nil {
		return fmt.Errorf("unable to delete bucket: %w", err)
	}

	if !globalstate.Quiet {
		fmt.Printf("Bucket %s%s deleted successfully\n", sos.BucketPrefix, c.Bucket)
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(storageCmd, &storageRbCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
