package storage

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type storageUploadCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"upload"`

	Files       []string `cli-arg:"#" cli-usage:"FILE"`
	Destination string   `cli-arg:"#" cli-usage:"sos://BUCKET/[PREFIX/]"`

	// The usage lists sos.ObjectCannedACLToStrings(): struct tags can't be
	// computed, update it if the SDK adds a canned ACL.
	ACL       string `cli-usage:"canned ACL to set on object (private|public-read|public-read-write|authenticated-read|aws-exec-read|bucket-owner-read|bucket-owner-full-control)"`
	DryRun    bool   `cli-short:"n" cli-usage:"simulate files upload, don't actually do it"`
	Recursive bool   `cli-short:"r" cli-usage:"upload directories recursively"`
}

func (c *storageUploadCmd) CmdAliases() []string { return []string{"put"} }
func (c *storageUploadCmd) CmdShort() string     { return "Upload files to a bucket" }
func (c *storageUploadCmd) CmdLong() string {
	return `This command uploads local files to a bucket.

Examples:

    # Upload files at the root of the bucket
    exo storage upload a b c sos://my-bucket

    # Upload files in a directory (trailing "/" in destination)
    exo storage upload index.html sos://my-bucket/public/

    # Upload a file under a different name
    exo storage upload a.txt sos://my-bucket/z.txt

    # Upload a directory recursively
    exo storage upload -r my-files/ sos://my-bucket
`
}

func (c *storageUploadCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	target := args[len(args)-1]
	args[len(args)-1] = strings.TrimPrefix(args[len(args)-1], sos.BucketPrefix)

	if bucket, _, _ := strings.Cut(args[len(args)-1], "/"); bucket == "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", target))
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if c.ACL != "" && !slices.Contains(sos.ObjectCannedACLToStrings(), c.ACL) {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid canned ACL %q, supported values are: %s",
			c.ACL, strings.Join(sos.ObjectCannedACLToStrings(), ", ")))
	}

	return nil
}

func (c *storageUploadCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var prefix string

	dstParts := strings.SplitN(c.Destination, "/", 2)
	bucket := dstParts[0]
	if len(dstParts) > 1 {
		// Tricky case: if the user specifies "<bucket>/" as destination,
		// strings.SplitN()'s result slice contains an empty string as last
		// item: in this case we set the prefix as "/" to mean the root of
		// the bucket.
		if dstParts[len(dstParts)-1] == "" {
			prefix = "/"
		} else {
			prefix = dstParts[1]
		}
	} else {
		prefix = "/"
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	return storage.UploadFiles(exocmd.GContext, c.Files, &sos.StorageUploadConfig{
		Bucket:    bucket,
		Prefix:    prefix,
		ACL:       c.ACL,
		Recursive: c.Recursive,
		DryRun:    c.DryRun,
	})
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storageUploadCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
