package storage

import (
	"fmt"
	"path/filepath"
	"strings"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type storageDownloadCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"download"`

	Source      string `cli-arg:"#" cli-usage:"sos://BUCKET/[OBJECT|PREFIX/]"`
	Destination string `cli-arg:"?" cli-usage:"DESTINATION"`

	DryRun    bool `cli-short:"n" cli-usage:"simulate files download, don't actually do it"`
	Force     bool `cli-short:"f" cli-usage:"overwrite existing destination files"`
	Recursive bool `cli-short:"r" cli-usage:"download prefix recursively"`
}

func (c *storageDownloadCmd) CmdAliases() []string { return []string{"get"} }
func (c *storageDownloadCmd) CmdShort() string     { return "Download files from a bucket" }
func (c *storageDownloadCmd) CmdLong() string {
	return `This command downloads files from a bucket.

If no destination argument is provided, files will be stored into the current
directory.

Examples:

    # Download a single file
    exo storage download sos://my-bucket/file-a

    # Download a single file and rename it locally
    exo storage download sos://my-bucket/file-a file-z

    # Download a prefix recursively (creating the folder tree if needed)
    exo storage download -r sos://my-bucket/public/ /tmp/public/
`
}

func (c *storageDownloadCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	target := args[0]
	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

	if bucket, _, _ := strings.Cut(args[0], "/"); bucket == "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", target))
	}

	// Append implicit root prefix ("/") if only a bucket name is specified in the source
	if !strings.Contains(args[0], "/") {
		args[0] += "/"
	}

	// Only clean a destination that was given: an explicit "" becomes ".",
	// while a missing destination stays "" (current directory, object name).
	if len(args) == 2 {
		args[1] = filepath.Clean(args[1])
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *storageDownloadCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var prefix string

	parts := strings.SplitN(c.Source, "/", 2)
	bucket := parts[0]
	if len(parts) > 1 {
		prefix = parts[1]
		if prefix == "" {
			prefix = "/"
		}
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	objects := make([]*s3types.Object, 0)
	if err := storage.ForEachObject(exocmd.GContext, bucket, prefix, c.Recursive, func(o *s3types.Object) error {
		objects = append(objects, o)
		return nil
	}); err != nil {
		return fmt.Errorf("error listing objects: %s", err)
	}

	if c.DryRun {
		fmt.Println("[DRY-RUN]")
	}

	switch len(objects) {
	case 0:
		fmt.Printf("no objects exist at %q\n", prefix)
	case 1: // single file download (with optional rename)
		object := objects[0]

		err := storage.DownloadFile(
			exocmd.GContext,
			bucket,
			c.Destination,
			object,
			c.Force,
			c.DryRun,
		)
		if err != nil {
			return fmt.Errorf("failed to download single file: %w", err)
		}
	default:
		err := storage.DownloadFiles(
			exocmd.GContext,
			bucket,
			prefix,
			c.Source,
			c.Destination,
			objects,
			c.Force,
			c.DryRun,
		)

		if err != nil {
			return fmt.Errorf("batch file download failed: %w", err)
		}
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(storageCmd, &storageDownloadCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
