package cmd

import (
	"fmt"
	"os"

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
		bucket := args[0]

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		if recursive { // Remove all files stored in the bucket before deleting it
			objectsCh := make(chan string)

			go func() {
				defer close(objectsCh)

				for obj := range sosClient.ListObjectsV2(bucket, "", true, gContext.Done()) {
					if obj.Err != nil {
						fmt.Fprintf(os.Stderr, "error: %s: %s\n", obj.Key, obj.Err)
						os.Exit(1)
					}
					objectsCh <- obj.Key
				}
			}()

			for rmObjErr := range sosClient.RemoveObjectsWithContext(gContext, bucket, objectsCh) {
				if rmObjErr.Err != nil {
					fmt.Fprintf(os.Stderr, "error: %s: %s\n", rmObjErr.ObjectName, rmObjErr.Err)
					os.Exit(1)
				}
			}
		}

		if err = sosClient.RemoveBucket(bucket); err != nil {
			return err
		}

		if !gQuiet {
			fmt.Printf("Bucket %q deleted successfully\n", bucket)
		}

		return nil
	},
}

func init() {
	sosDeleteCmd.Flags().BoolP("recursive", "r", false, "Attempt to empty the bucket before deleting it")
	sosCmd.AddCommand(sosDeleteCmd)
}
