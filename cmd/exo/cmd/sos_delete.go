package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// sosDeleteCmd represents the delete command
var sosDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Short:   "Delete a bucket",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			log.Fatal(err)
		}

		zone, err := minioClient.GetBucketLocation(args[0])
		if err != nil {
			return err
		}

		minioClient, err = newMinioClient(zone)
		if err != nil {
			return err
		}

		return minioClient.RemoveBucket(args[0])
	},
}

func init() {
	sosCmd.AddCommand(sosDeleteCmd)
}
