package storage

import (
	"fmt"

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

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if _, key, err := parsePath(c.Bucket); err != nil {
		clicmd.CmdExitOnUsageError(cmd, err.Error())
	} else if key != "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q, expected a bucket name", c.Bucket))
	}

	return nil
}

func (c *storageRbCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket, _, err := parsePath(c.Bucket)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf("Are you sure you want to delete %s%s?", sos.BucketPrefix, bucket)) {
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

	if err := storage.DeleteBucket(exocmd.GContext, bucket, c.Recursive); err != nil {
		return fmt.Errorf("unable to delete bucket: %w", err)
	}

	if !globalstate.Quiet {
		fmt.Printf("Bucket %s%s deleted successfully\n", sos.BucketPrefix, bucket)
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storageRbCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
