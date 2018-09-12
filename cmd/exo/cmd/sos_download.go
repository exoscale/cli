package cmd

import (
	"log"

	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <bucket name> <object name> <file path>",
	Short: "Download an object from a bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return cmd.Usage()
		}

		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			return err
		}

		location, err := minioClient.GetBucketLocation(args[0])
		if err != nil {
			return err
		}

		minioClient, err = newMinioClient(location)
		if err != nil {
			return err
		}

		if err = minioClient.FGetObjectWithContext(gContext, args[0], args[1], args[2], minio.GetObjectOptions{}); err != nil {
			return err
		}

		log.Printf("Successfully downloaded %s into %q\n", args[1], args[2])
		return nil
	},
}

func init() {
	sosCmd.AddCommand(downloadCmd)
}
