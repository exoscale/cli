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

		//It use asyncTasks to have spinner when user exec this command
		r := asyncTasks([]task{task{
			egoscale.UpdateInstancePool{
				ID:          instancePool.ID,
				Description: description,
				ZoneID:      zone.ID,
				Userdata:    userData,
			},
			fmt.Sprintf("Update instance pool %q", args[0]),
		}})
		errs := filterErrors(r)
		if len(errs) > 0 {
			return errs[0]
		}

		return nil
	},
}

func init() {
	instancePoolUpdateCmd.Flags().StringP("description", "d", "", "Instance pool description")
	instancePoolUpdateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init file path")
	instancePoolCmd.AddCommand(instancePoolUpdateCmd)
}
