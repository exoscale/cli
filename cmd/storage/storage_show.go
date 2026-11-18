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

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if _, _, err := parsePath(c.Target); err != nil {
		clicmd.CmdExitOnUsageError(cmd, err.Error())
	}

	return nil
}

func (c *storageShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket, key, err := parsePath(c.Target)
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

	if key == "" {
		return utils.PrintOutput(storage.ShowBucket(exocmd.GContext, bucket))
	}

	return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, key))
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storageShowCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
