package cmd

import (
	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var sosCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "create bucket",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zone != "" {
			gCurrentAccount.DefaultZone = zone
		}

		minioClient, err := newMinioClient(gCurrentAccount.DefaultZone)
		if err != nil {
			return err
		}

		return createBucket(minioClient, args[0], gCurrentAccount.DefaultZone)
	},
}

func createBucket(minioClient *minio.Client, bucketName, zone string) error {
	return minioClient.MakeBucket(bucketName, zone)
}

func init() {
	sosCmd.AddCommand(sosCreateCmd)
	sosCreateCmd.Flags().StringP("zone", "z", "", "Simple object storage zone")
}
