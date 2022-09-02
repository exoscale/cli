package cmd

import (
	"github.com/spf13/cobra"
)

var sosCreateCmd = &cobra.Command{
	Use:     "create NAME",
	Short:   "Create a bucket",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		bucket := args[0]

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		if zone != "" {
			if err := sosClient.setZone(zone); err != nil {
				return err
			}
		}

		return createBucket(sosClient, bucket, zone)
	},
}

func createBucket(sosClient *sosClient, name, zone string) error {
	return sosClient.MakeBucket(name, zone)
}

func init() {
	sosCmd.AddCommand(sosCreateCmd)
	sosCreateCmd.Flags().StringP("zone", "z", "", "Simple object storage zone")
}
