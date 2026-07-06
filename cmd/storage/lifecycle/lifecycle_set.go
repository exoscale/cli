package lifecycle

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type setCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"set"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`
	File   string `cli-arg:"#" cli-usage:"path/to/lifecycle.json"`

	Zone string `cli-short:"z" cli-usage:"zone"`
}

func (c *setCmd) CmdAliases() []string { return nil }
func (c *setCmd) CmdShort() string     { return "Set lifecycle configuration" }
func (c *setCmd) CmdLong() string {
	return `Set a lifecycle configuration for a bucket.

Example of a valid lifecycle configuration:
{
    "Rules": [
        {
            "Status": "Enabled",
            "Expiration": { "Days": 30 },
            "Filter": { "Prefix": "" },
            "ID": "expire-after-30-days"
        }
    ]
}`
}

func (c *setCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *setCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	confFile, err := os.Open(c.File)
	if err != nil {
		return err
	}
	defer func() { _ = confFile.Close() }()

	var configuration sos.BucketLifecycleConf
	if err = json.NewDecoder(confFile).Decode(&configuration); err != nil {
		return err
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptWithZone(c.Zone),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	return storage.PutBucketLifecycle(exocmd.GContext, bucket, configuration.ToS3())
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &setCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
