package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var instancePoolUpdateCmd = &cobra.Command{
	Use:     "update <name>",
	Short:   "Update an instance pool",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		zone, err := getZoneByName(gCurrentAccount.DefaultZone)
		if err != nil {
			return err
		}

		userDataPath, err := cmd.Flags().GetString("cloud-init")
		if err != nil {
			return err
		}

		userData := ""
		if userDataPath != "" {
			userData, err = getUserData(userDataPath)
			if err != nil {
				return err
			}

			if len(userData) >= maxUserDataLength {
				return fmt.Errorf("user-data maximum allowed length is %d bytes", maxUserDataLength)
			}
		}

		instancePool, err := getInstancePoolByName(args[0], zone.ID)
		if err != nil {
			return err
		}

		size, err := cmd.Flags().GetInt("size")
		if err != nil {
			return err
		}

		_, err = cs.RequestWithContext(gContext, &egoscale.UpdateInstancePool{
			ID:          instancePool.ID,
			Description: description,
			ZoneID:      zone.ID,
			UserData:    userData,
		})
		if err != nil {
			return err
		}

		if size > 0 {
			_, err = cs.RequestWithContext(gContext, &egoscale.ScaleInstancePool{
				ID:     instancePool.ID,
				ZoneID: instancePool.ZoneID,
				Size:   size,
			})
			if err != nil {
				return err
			}
		}

		return showInstancePool(instancePool.Name)
	},
}

func init() {
	instancePoolUpdateCmd.Flags().StringP("description", "d", "", "Instance pool description")
	instancePoolUpdateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init file path")
	instancePoolUpdateCmd.Flags().IntP("size", "s", 0, "Update instance pool size")
	instancePoolCmd.AddCommand(instancePoolUpdateCmd)
}
