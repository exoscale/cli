package storage

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type storagePurgeCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"purge"`

	Target string `cli-arg:"#" cli-usage:"sos://BUCKET/[OBJECT|PREFIX/]"`

	Force   bool `cli-short:"f" cli-usage:"attempt to perform the operation without prompting for confirmation"`
	Verbose bool `cli-short:"v" cli-usage:"output deleted objects"`
}

func (c *storagePurgeCmd) CmdAliases() []string { return nil }
func (c *storagePurgeCmd) CmdShort() string     { return "Purge objects and their versions" }
func (c *storagePurgeCmd) CmdLong() string {
	return `This command purges objects and object versions stored in a bucket.

    exo storage purge sos://my-bucket/
    exo storage purge sos://my-bucket/some-prefix/
`
}

func (c *storagePurgeCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if _, _, err := parsePath(c.Target); err != nil {
		clicmd.CmdExitOnUsageError(cmd, err.Error())
	}

	return nil
}

func (c *storagePurgeCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket, prefix, err := parsePath(c.Target)
	if err != nil {
		return err
	}

	// Special case: the caller wants to target objects at the root of the
	// bucket, in this case the prefix is empty so we set it to a symbolic
	// value that shall be removed later on.
	if prefix == "" {
		prefix = "/"
	}

	if !c.Force {
		if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf("Are you sure you want to delete %s%s/%s?",
			sos.BucketPrefix, bucket, prefix)) {
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

	deletedChan, errChan := storage.DeleteObjectVersions(exocmd.GContext, bucket, prefix)

	for deletedChan != nil || errChan != nil {
		select {
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			fmt.Printf("Error: %v\n", err)
		case deletedElt, ok := <-deletedChan:
			if !ok {
				deletedChan = nil
				continue
			}
			if c.Verbose {
				fmt.Println("deleted:", aws.ToString(deletedElt.Key))
			}
		}
	}
	fmt.Println("Purge completed")
	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storagePurgeCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
