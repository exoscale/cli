package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// templateRegisterCmd registers a template
var templateRegisterCmd = &cobra.Command{
	Use:     "register",
	Short:   "register a custom template",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		checksum, err := cmd.Flags().GetString("checksum")
		if err != nil {
			return err
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zone == "" {
			zone = gCurrentAccount.DefaultZone
		}

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			return err
		}

		disablePassword, err := cmd.Flags().GetBool("disable-password")
		if err != nil {
			return err
		}

		enablePassword := !(disablePassword)

		disableSSHKey, err := cmd.Flags().GetBool("disable-ssh-key")
		if err != nil {
			return err
		}

		enableSSHKey := !(disableSSHKey)

		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}

		details := make(map[string]string)
		details["username"] = username

		req := egoscale.RegisterCustomTemplate{
			Checksum:        checksum,
			Details:         details,
			Displaytext:     description,
			Name:            name,
			PasswordEnabled: &enablePassword,
			SSHKeyEnabled:   &enableSSHKey,
			URL:             url,
		}

		return templateRegister(req, zone)
	},
}

func templateRegister(registerTemplate egoscale.RegisterCustomTemplate, zone string) error {
	zoneID, err := getZoneIDByName(zone)
	if err != nil {
		return err
	}
	registerTemplate.ZoneID = zoneID

	resp, err := asyncRequest(registerTemplate, fmt.Sprintf("Registering the template"))
	if err != nil {
		return err
	}

	templates := resp.(*[]egoscale.Template)

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Zone", "Name", "ID"})
	for _, template := range *templates {
		table.Append([]string{template.ZoneName, template.Name, template.ID.String()})
	}

	table.Render()

	return nil
}

func init() {
	templateRegisterCmd.Flags().StringP("checksum", "", "", "the MD5 checksum value of the template")
	templateRegisterCmd.Flags().StringP("description", "", "", "the template description")
	templateRegisterCmd.Flags().StringP("name", "", "", "the name of the template")
	templateRegisterCmd.Flags().StringP("zone", "", "", "the ID or name of the zone the template is to be hosted on")
	templateRegisterCmd.Flags().StringP("url", "", "", "the URL of where the template is hosted")
	templateRegisterCmd.Flags().BoolP("disable-password", "", false, "true if the template does not support password authentication; default is false")
	templateRegisterCmd.Flags().BoolP("disable-ssh-key", "", false, "true if the template does not support ssh key authentication; default is false")
	templateRegisterCmd.Flags().StringP("username", "", "", "The default username of the template")

	templateCmd.AddCommand(templateRegisterCmd)
}
