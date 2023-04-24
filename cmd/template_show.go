package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/output"
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
	BootMode     string `json:"boot_mode"`
}

func (o *templateShowOutput) Type() string { return "Template" }
func (o *templateShowOutput) toJSON()      { output.JSON(o) }
func (o *templateShowOutput) toText()      { output.Text(o) }
func (o *templateShowOutput) toTable()     { output.Table(o) }

func init() {
	templateShowCmd := &cobra.Command{
		Use:   "show NAME|ID",
		Short: "Show a template details",
		Long: fmt.Sprintf(`This command shows a Compute instance template details.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&templateShowOutput{}), ", ")),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmdExitOnUsageError(cmd, "invalid arguments")
			}

			cmdSetZoneFlagFromDefault(cmd)

			return cmdCheckRequiredFlags(cmd, []string{"zone"})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			zone, err := getZoneByNameOrID(zoneName)
			if err != nil {
				return err
			}

			template, err := getTemplateByNameOrID(zone.ID, name, templateFilter)
			if err != nil {
				return err
			}

			return printOutput(showTemplate(template))
		},
	}

	templateShowCmd.Flags().StringP("template-filter", "", defaultTemplateFilter, templateFilterHelp)
	templateShowCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneHelp)
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
		BootMode:     template.BootMode,
	}

	if username, ok := template.Details["username"]; ok {
		out.Username = username
	}

	return &out, nil
}
