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

type addCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	Target    string   `cli-arg:"#" cli-usage:"sos://BUCKET/(OBJECT|PREFIX/)"`
	KeyValues []string `cli-arg:"#" cli-usage:"KEY=VALUE"`

	Recursive bool `cli-short:"r" cli-usage:"add metadata recursively (with object prefix only)"`
}

func (c *addCmd) CmdAliases() []string { return nil }
func (c *addCmd) CmdShort() string     { return "Add key/value metadata to an object" }
func (c *addCmd) CmdLong() string {
	return fmt.Sprintf(`This command adds key/value metadata to an object.

Example:

    exo storage metadata add sos://my-bucket/object-a \
        k1=v1 \
        k2=v2

Notes:

  * Adding an already existing key will overwrite its value.
  * The following characters are not allowed in keys: %s

Supported output template annotations: %s`,
		strings.Join(strings.Split(sos.MetadataForbiddenCharset, ""), " "),
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", "))
}

func (c *addCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

	if !strings.Contains(args[0], "/") {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
	}

	for _, kv := range args[1:] {
		if !strings.Contains(kv, "=") {
			clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", kv))
		}
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addCmd) CmdRun(_ *cobra.Command, _ []string) error {
	metadata := make(map[string]string)

	parts := strings.SplitN(c.Target, "/", 2)
	bucket, prefix := parts[0], parts[1]

	for _, kv := range c.KeyValues {
		parts = strings.Split(kv, "=")
		metadata[parts[0]] = parts[1]
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	if err := storage.AddObjectsMetadata(exocmd.GContext, bucket, prefix, metadata, c.Recursive); err != nil {
		return fmt.Errorf("unable to add metadata to object: %w", err)
	}

	if !globalstate.Quiet && !c.Recursive && !strings.HasSuffix(prefix, "/") {
		return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
	}

	if !globalstate.Quiet {
		fmt.Println("Metadata added successfully")
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
