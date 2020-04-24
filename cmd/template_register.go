package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// templateRegisterCmd registers a template
var templateRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "register a custom template",
	Long: fmt.Sprintf(`This command registers a new Compute instance template.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&templateShowOutput{}), ", ")),
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		} else if name == "" {
			return fmt.Errorf("template name must be specified")
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		} else if description == "" {
			return fmt.Errorf("template description must be specified")
		}

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			return err
		} else if url == "" {
			return fmt.Errorf("template image URL must be specified")
		}

		checksum, err := cmd.Flags().GetString("checksum")
		if err != nil {
			return err
		} else if checksum == "" {
			return fmt.Errorf("template image file checksum must be specified")
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zone == "" {
			zone = gCurrentAccount.DefaultZone
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

		req := egoscale.RegisterCustomTemplate{
			Name:            name,
			URL:             url,
			Checksum:        checksum,
			Displaytext:     description,
			PasswordEnabled: &enablePassword,
			SSHKeyEnabled:   &enableSSHKey,
		}

		if username, _ := cmd.Flags().GetString("username"); username != "" {
			req.Details = make(map[string]string)
			req.Details["username"] = username
		}

		return templateRegister(req, zone)
	},
}

func templateRegister(registerTemplate egoscale.RegisterCustomTemplate, zone string) error {
	z, err := getZoneByName(zone)
	if err != nil {
		return err
	}
	registerTemplate.ZoneID = z.ID

	resp, err := asyncRequest(registerTemplate, "Registering the template")
	if err != nil {
		return err
	}

	templates := resp.(*[]egoscale.Template)
	if len(*templates) != 1 {
		return nil
	}
	template := (*templates)[0]

	if !gQuiet {
		return output(showTemplate(&template))
	}

	return nil
}

func init() {
	templateRegisterCmd.Flags().StringP("checksum", "", "", "the MD5 checksum value of the template")
	templateRegisterCmd.Flags().StringP("description", "", "", "the template description")
	templateRegisterCmd.Flags().StringP("name", "", "", "the name of the template")
	templateRegisterCmd.Flags().StringP("zone", "z", "", "the ID or name of the zone the template is to be hosted on")
	templateRegisterCmd.Flags().StringP("url", "", "", "the URL of where the template is hosted")
	templateRegisterCmd.Flags().BoolP("disable-password", "", false, "true if the template does not support password authentication; default is false")
	templateRegisterCmd.Flags().BoolP("disable-ssh-key", "", false, "true if the template does not support ssh key authentication; default is false")
	templateRegisterCmd.Flags().StringP("username", "", "", "The default username of the template")

	templateCmd.AddCommand(templateRegisterCmd)
}
