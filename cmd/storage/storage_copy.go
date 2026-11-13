package storage

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type storageCopyCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"copy"`

	Source      string `cli-arg:"#" cli-usage:"sos://BUCKET/[OBJECT|PREFIX/]"`
	Destination string `cli-arg:"#" cli-usage:"sos://BUCKET/[OBJECT|PREFIX/]"`

	DryRun               bool  `cli-short:"n" cli-usage:"simulate the copy operation"`
	Force                bool  `cli-short:"f" cli-usage:"skip confirmation prompt"`
	Recursive            bool  `cli-short:"r" cli-usage:"copy objects recursively"`
	Verbose              bool  `cli-short:"v" cli-usage:"output copied objects"`
	MultipartConcurrency int64 `cli-usage:"number of concurrent parts for multipart copies"`
}

func (c *storageCopyCmd) CmdAliases() []string { return []string{"cp"} }
func (c *storageCopyCmd) CmdShort() string     { return "Copy objects within a bucket or across buckets" }
func (c *storageCopyCmd) CmdLong() string {
	return `Copy objects within a bucket or across buckets.

This command copies objects server-side without downloading them. Object
metadata, headers, and ACLs are preserved. Existing destination objects are
overwritten.

Multi-object prefix copies are processed serially. A trailing slash on the
source selects prefix mode; -r controls recursion into subdirectories.

Examples:

    exo storage copy sos://my-bucket/file-a sos://my-bucket/folder/file-a

    exo storage copy sos://my-bucket/file-a sos://other-bucket/file-a

    exo storage copy -r sos://my-bucket/prefix/ sos://other-bucket/prefix/

    exo storage copy -n sos://my-bucket/file-a sos://other-bucket/file-a
`
}

func (c *storageCopyCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	return validateStorageTransferArgs(c.Source, c.Destination, c.Recursive)
}

func (c *storageCopyCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	return runStorageTransfer(cmd.Name(), c.Source, c.Destination, storageTransferOptions{
		Recursive:            c.Recursive,
		Force:                c.Force,
		Verbose:              c.Verbose,
		DryRun:               c.DryRun,
		MultipartConcurrency: int(c.MultipartConcurrency),
	}, false)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(storageCmd, &storageCopyCmd{
		CliCommandSettings:   clicmd.DefaultCLICmdSettings(),
		MultipartConcurrency: 4,
	}))
}
