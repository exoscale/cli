package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

var templateShowCmd = &cobra.Command{
	Use:   "show <template name | id>",
	Short: "Show a template details",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		name := args[0]

		templateFilterCmd, err := cmd.Flags().GetString("template-filter")
		if err != nil {
			return err
		}
		templateFilter, err := validateTemplateFilter(templateFilterCmd)
		if err != nil {
			return err
		}

		zoneName, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zoneName == "" {
			zoneName = gCurrentAccount.DefaultZone
		}

		zoneID, err := getZoneIDByName(zoneName)
		if err != nil {
			return err
		}

		template, err := getTemplateByName(zoneID, name, templateFilter)
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
	},
}

func init() {
	templateCmd.AddCommand(templateShowCmd)
	templateShowCmd.Flags().StringP("template-filter", "", "featured", "The template filter to use (mine,community,featured)")
	templateShowCmd.Flags().StringP("zone", "z", "", zoneHelp)
}
