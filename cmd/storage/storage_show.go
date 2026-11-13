package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type storageShowCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Target string `cli-arg:"#" cli-usage:"sos://BUCKET/[OBJECT]"`
}

func (c *storageShowCmd) CmdAliases() []string { return nil }
func (c *storageShowCmd) CmdShort() string     { return "Show a bucket/object details" }
func (c *storageShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows the details of a Storage bucket or object.

Supported output template annotations:

	* When showing a bucket: %s
	* When showing an object: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowBucketOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", "))
}

func (c *storageShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	target := args[0]
	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

	if bucket, _, _ := strings.Cut(args[0], "/"); bucket == "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", target))
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *storageShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var key string

	parts := strings.SplitN(c.Target, "/", 2)
	bucket := parts[0]
	if len(parts) > 1 {
		key = parts[1]
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	if key == "" {
		return utils.PrintOutput(storage.ShowBucket(exocmd.GContext, bucket))
	}

	return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, key))
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(storageCmd, &storageShowCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
