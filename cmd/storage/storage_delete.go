package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type storageDeleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Target string `cli-arg:"#" cli-usage:"sos://BUCKET/[OBJECT|PREFIX/]"`

	Force     bool `cli-short:"f" cli-usage:"attempt to perform the operation without prompting for confirmation"`
	Recursive bool `cli-short:"r" cli-usage:"delete objects recursively"`
	Verbose   bool `cli-short:"v" cli-usage:"output deleted objects"`
}

func (c *storageDeleteCmd) CmdAliases() []string { return []string{"del", "rm"} }
func (c *storageDeleteCmd) CmdShort() string     { return "Delete objects" }
func (c *storageDeleteCmd) CmdLong() string {
	return `This command deletes objects stored in a bucket.

If you want to target objects under a "directory" prefix, suffix the path
argument with "/":

    exo storage delete sos://my-bucket/
    exo storage delete -r sos://my-bucket/some-directory/
`
}

func (c *storageDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

	if !strings.Contains(args[0], "/") {
		args[0] += "/"
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *storageDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	deleted, err := storage.DeleteObjects(exocmd.GContext, bucket, prefix, c.Recursive)
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			// Error in individual files, print to stderr & continue
			for _, e := range merr.Errors {
				fmt.Fprintln(os.Stderr, e)
			}
		} else {
			// Global error, exit
			return fmt.Errorf("unable to delete objects: %w", err)
		}
	}

	if c.Verbose {
		for _, o := range deleted {
			fmt.Println(aws.ToString(o.Key))
		}
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(storageCmd, &storageDeleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
