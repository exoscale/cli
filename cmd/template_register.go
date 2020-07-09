package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// templateRegisterCmd registers a template
var templateRegisterCmd = &cobra.Command{
	Use:   "register <name>",
	Short: "register a custom template",
	Long: fmt.Sprintf(`This command registers a new Compute instance template.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&templateShowOutput{}), ", ")),
	Aliases: gCreateAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		if cmd.Flags().Changed("from-snapshot") {
			return cmdCheckRequiredFlags(cmd, []string{
				"zone",
			})
		}

		return cmdCheckRequiredFlags(cmd, []string{
			"zone",
			"url",
			"checksum",
		})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			return err
		}

		checksum, err := cmd.Flags().GetString("checksum")
		if err != nil {
			return err
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		username, err := cmd.Flags().GetString("username")
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

		bootmode, err := cmd.Flags().GetString("boot-mode")
		if err != nil {
			return err
		}

		snapshotID, err := cmd.Flags().GetString("from-snapshot")
		if err != nil {
			return err
		}

		req := egoscale.RegisterCustomTemplate{
			Name:            args[0],
			URL:             url,
			Checksum:        checksum,
			Displaytext:     description,
			PasswordEnabled: &enablePassword,
			SSHKeyEnabled:   &enableSSHKey,
			BootMode:        bootmode,
		}

		if username != "" {
			req.Details = make(map[string]string)
			req.Details["username"] = username
		}

		if snapshotID != "" {
			snapshot, err := exportSnapshot(snapshotID)
			if err != nil {
				return err
			}

			req.Checksum = snapshot.MD5sum
			req.URL = snapshot.PresignedURL
		}

		return output(templateRegister(req, zone))
	},
}

func templateRegister(registerTemplate egoscale.RegisterCustomTemplate, zone string) (outputter, error) {
	z, err := getZoneByName(zone)
	if err != nil {
		return nil, err
	}
	registerTemplate.ZoneID = z.ID

	resp, err := asyncRequest(registerTemplate, "Registering the template")
	if err != nil {
		return nil, err
	}

	templates := resp.(*[]egoscale.Template)
	if len(*templates) != 1 {
		return nil, nil
	}
	template := (*templates)[0]

	if !gQuiet {
		return showTemplate(&template)
	}

	return nil, nil
}

func init() {
	templateRegisterCmd.Flags().StringP("checksum", "c", "", "MD5 checksum value of the template")
	templateRegisterCmd.Flags().StringP("description", "d", "", "template description")
	templateRegisterCmd.Flags().StringP("zone", "z", "", "ID or name of the zone the template is to be hosted on")
	templateRegisterCmd.Flags().StringP("username", "u", "", "default username of the template")
	templateRegisterCmd.Flags().String("url", "", "URL of where the template is hosted")
	templateRegisterCmd.Flags().String("boot-mode", "legacy", "template boot mode (legacy|uefi)")
	templateRegisterCmd.Flags().String("from-snapshot", "", "ID of a Compute instance snapshot to register as a new template")
	templateRegisterCmd.Flags().Bool("disable-password", false, "disable password-based authentication")
	templateRegisterCmd.Flags().Bool("disable-ssh-key", false, "disable SSH key-based authentication")

	templateCmd.AddCommand(templateRegisterCmd)
}
