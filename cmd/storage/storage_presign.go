package storage

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type storagePresignCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"presign"`

	Object string `cli-arg:"#" cli-usage:"sos://BUCKET/OBJECT"`

	Expires time.Duration `cli-short:"e" cli-usage:"expiration duration for the generated pre-signed URL (e.g. \"1h45m\", \"30s\"); supported units: \"s\", \"m\", \"h\""`
	Method  string        `cli-short:"m" cli-usage:"pre-signed URL method (get|put)"`
}

func (c *storagePresignCmd) CmdAliases() []string { return nil }
func (c *storagePresignCmd) CmdShort() string     { return "Generate a pre-signed URL to an object" }
func (c *storagePresignCmd) CmdLong() string      { return "" }

func (c *storagePresignCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	// Same usage error (message, usage and exit code) as the legacy command.
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}
	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if _, key, err := parsePath(c.Object); err != nil || key == "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid object URL: %q", c.Object))
	}

	return nil
}

func (c *storagePresignCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket, key, err := parsePath(c.Object)
	if err != nil {
		return err
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	url, err := storage.GenPresignedURL(exocmd.GContext, c.Method, bucket, key, c.Expires)
	if err != nil {
		return fmt.Errorf("unable to pre-sign %s%s/%s: %w", sos.BucketPrefix, bucket, key, err)
	}

	fmt.Println(url)

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storagePresignCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		Expires:            900 * time.Second,
		Method:             "get",
	}))
}
