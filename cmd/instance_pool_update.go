package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var instancePoolUpdateCmd = &cobra.Command{
	Use:   "update <name | id>",
	Short: "Update an instance pool",
	Long: fmt.Sprintf(`This command updates an instance pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolItemOutput{}), ", ")),
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		zoneflag, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		zone, err := getZoneByName(zoneflag)
		if err != nil {
			return err
		}

		templateFilterCmd, err := cmd.Flags().GetString("template-filter")
		if err != nil {
			return err
		}
		templateFilter, err := validateTemplateFilter(templateFilterCmd)
		if err != nil {
			return err
		}

		template := new(egoscale.Template)
		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		if templateName != "" {
			template, err = getTemplateByName(zone.ID, templateName, templateFilter)
			if err != nil {
				return err
			}
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
			ZoneID:      zone.ID,
			Name:        name,
			Description: description,
			TemplateID:  template.ID,
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

		if !gQuiet {
			return showInstancePool(instancePool.ID.String(), instancePool.ZoneID.String())
		}

		return nil
	},
}

func init() {
	// Required Flags
	instancePoolUpdateCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	if err := instancePoolUpdateCmd.MarkFlagRequired("zone"); err != nil {
		log.Fatal(err)
	}

	instancePoolUpdateCmd.Flags().StringP("name", "n", "", "Instance pool name")
	instancePoolUpdateCmd.Flags().StringP("description", "d", "", "Instance pool description")
	instancePoolUpdateCmd.Flags().StringP("template", "t", "", "Instance pool template")
	instancePoolUpdateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolUpdateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init file path")
	instancePoolUpdateCmd.Flags().IntP("size", "s", 0, "Update instance pool size")
	instancePoolCmd.AddCommand(instancePoolUpdateCmd)
}
