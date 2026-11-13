package metadata

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type deleteCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Target string   `cli-arg:"#" cli-usage:"sos://BUCKET/(OBJECT|PREFIX/)"`
	Keys   []string `cli-arg:"#" cli-usage:"KEY"`

	Recursive bool `cli-short:"r" cli-usage:"delete metadata recursively (with object prefix only)"`
}

func (c *deleteCmd) CmdAliases() []string { return []string{"del"} }
func (c *deleteCmd) CmdShort() string     { return "Delete metadata from an object" }
func (c *deleteCmd) CmdLong() string {
	return fmt.Sprintf(`This command deletes key/value metadata from an object.

Example:

    exo storage metadata delete sos://my-bucket/object-a k1

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", "))
}

func (c *deleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

	if !strings.Contains(args[0], "/") {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *deleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	parts := strings.SplitN(c.Target, "/", 2)
	bucket, prefix := parts[0], parts[1]

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	if err := storage.DeleteObjectsMetadata(exocmd.GContext, bucket, prefix, c.Keys, c.Recursive); err != nil {
		return fmt.Errorf("unable to delete metadata from object: %w", err)
	}

	if !globalstate.Quiet && !c.Recursive && !strings.HasSuffix(prefix, "/") {
		return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
	}

	if !globalstate.Quiet {
		fmt.Println("Metadata deleted successfully")
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &deleteCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
