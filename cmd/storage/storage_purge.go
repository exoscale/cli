package storage

import (
	"fmt"
	"strings"

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

	target := args[0]
	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

	if bucket, _, _ := strings.Cut(args[0], "/"); bucket == "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", target))
	}

	if !strings.Contains(args[0], "/") {
		args[0] += "/"
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *storagePurgeCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var prefix string

	parts := strings.SplitN(c.Target, "/", 2)
	bucket := parts[0]
	if len(parts) > 1 {
		prefix = parts[1]

		// Special case: the caller wants to target objects at the root of
		// the bucket, in this case the prefix is empty so we set the key
		// to a symbolic value that shall be removed later on.
		if prefix == "" {
			prefix = "/"
		}
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
