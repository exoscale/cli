package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type templateShowOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	OSType       string `json:"os_type" outputLabel:"OS Type"`
	CreationDate string `json:"creation_date"`
	Zone         string `json:"zone"`
	DiskSize     string `json:"disk_size"`
	Username     string `json:"username"`
	Password     bool   `json:"password" outputLabel:"Password?"`
}

func (o *templateShowOutput) Type() string { return "Template" }
func (o *templateShowOutput) toJSON()      { outputJSON(o) }
func (o *templateShowOutput) toText()      { outputText(o) }
func (o *templateShowOutput) toTable()     { outputTable(o) }

func init() {
	var templateShowCmd = &cobra.Command{
		Use:   "show <template name | id>",
		Short: "Show a template details",
		Long: fmt.Sprintf(`This command shows a Compute instance template details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&templateShowOutput{}), ", ")),
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

			zone, err := getZoneByName(zoneName)
			if err != nil {
				return err
			}

			template, err := getTemplateByName(zone.ID, name, templateFilter)
			if err != nil {
				return err
			}

			return output(showTemplate(template))
		},
	}

	templateShowCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	templateShowCmd.Flags().StringP("zone", "z", "", zoneHelp)
	templateCmd.AddCommand(templateShowCmd)
}

func showTemplate(template *egoscale.Template) (outputter, error) {
	out := templateShowOutput{
		ID:           template.ID.String(),
		Name:         template.Name,
		OSType:       template.OsTypeName,
		CreationDate: template.Created,
		Zone:         template.ZoneName,
		DiskSize:     humanize.IBytes(uint64(template.Size)),
		Password:     template.PasswordEnabled,
	}

	if username, ok := template.Details["username"]; ok {
		out.Username = username
	}

	return &out, nil
}
