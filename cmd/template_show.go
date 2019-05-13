package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

func init() {
	templateCmd.AddCommand(templateShowCmd)
}

// templateShowCmd represents the show command
var templateShowCmd = &cobra.Command{
	Use:   "show <template name | id>",
	Short: "Show a template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("show expects one template by name or id")
		}
		return showTemplate(args[0])
	},
}

func showTemplate(name string) error {
	zoneID, err := getZoneIDByName(gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	template, err := getTemplateByName(zoneID, name)
	if err != nil {
		return err
	}

	username, usernameOk := template.Details["username"]

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{template.Name})

	t.Append([]string{"ID", template.ID.String()})
	t.Append([]string{"Name", template.Name})
	t.Append([]string{"OS Type", template.OsTypeName})
	t.Append([]string{"Created", template.Created})
	t.Append([]string{"Size", fmt.Sprintf("%d", template.Size>>30)})

	if usernameOk {
		t.Append([]string{"Username", username})
	}

	t.Append([]string{"Password?", fmt.Sprintf("%t", template.PasswordEnabled)})

	t.Render()

	return nil
}
