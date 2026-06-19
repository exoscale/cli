package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketLifecycleSetCmd.Flags().StringP(exocmd.ZoneFlagLong, exocmd.ZoneFlagShort, "", exocmd.ZoneFlagMsg)
	storageBucketLifecycleCmd.AddCommand(storageBucketLifecycleSetCmd)
}

var storageBucketLifecycleSetCmd = &cobra.Command{
	Use:   "set sos://BUCKET path/to/lifecycle.json",
	Short: "Set lifecycle configuration",
	Long: `Set a lifecycle configuration for a bucket.

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
}`,
	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		exocmd.CmdSetZoneFlagFromDefault(cmd)
		return exocmd.CmdCheckRequiredFlags(cmd, []string{exocmd.ZoneFlagLong})
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		var configuration sos.BucketLifecycleConf

		confFile, err := os.Open(args[1])
		if err != nil {
			return err
		}

		jsonParsr := json.NewDecoder(confFile)
		if err = jsonParsr.Decode(&configuration); err != nil {
			return err
		}

		bucket := args[0]

		zone, err := cmd.Flags().GetString(exocmd.ZoneFlagLong)
		if err != nil {
			return err
		}

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		s3conf := configuration.ToS3()

		err = storage.PutBucketLifecycle(exocmd.GContext, bucket, s3conf)

		return err
	},
}
