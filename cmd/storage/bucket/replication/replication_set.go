package replication

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

type setCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"set"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`
	File   string `cli-arg:"#" cli-usage:"path/to/replication.json"`

	Zone string `cli-short:"z" cli-usage:"bucket zone"`
}

func (c *setCmd) CmdAliases() []string { return nil }
func (c *setCmd) CmdShort() string     { return "set replication configuration" }
func (c *setCmd) CmdLong() string {
	return `Set a replication configuration for a bucket. Bucket versioning needs to be enabled
for both source and target bucket

Example of a valid replication configuration:
{
    "Role": "role-uuid",
    "Rules": [
        {
            "Status": "Enabled",
            "Priority": 1,
            "DeleteMarkerReplication": { "Status": "Disabled" },
            "Filter" : { "Prefix": ""},
            "Destination": {
                "Bucket": "target-bucket"
            },
            "ID": "foo"
        }
    ]
}

More information at https://docs.aws.amazon.com/cli/latest/reference/s3api/put-bucket-replication.html#options & https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketReplication.html`
}

func (c *setCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if err := checkArgsAndZone(cmd, args, 2); err != nil {
		return err
	}
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *setCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	confFile, err := os.Open(c.File)
	if err != nil {
		return err
	}
	defer func() { _ = confFile.Close() }()

	var configuration sos.BucketReplicationConf
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

	return storage.PutBucketReplication(exocmd.GContext, bucket, configuration.ToS3())
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &setCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
