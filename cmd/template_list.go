package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/output"
	"github.com/spf13/cobra"
)

type templateListItemOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
	Zone         string `json:"zone"`
	DiskSize     string `json:"disk_size"`
}

type templateListOutput []templateListItemOutput

func (o *templateListOutput) toJSON()  { output.JSON(o) }
func (o *templateListOutput) toText()  { output.Text(o) }
func (o *templateListOutput) toTable() { output.Table(o) }

func init() {
	templateListCmd.Flags().BoolP("community", "", false, "List community templates")
	templateListCmd.Flags().BoolP("featured", "", false, "List featured templates")
	templateListCmd.Flags().BoolP("mine", "", false, "List your templates")
	templateListCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", "Name of the zone (default: current account's default zone)")
	templateCmd.AddCommand(templateListCmd)
}

var templateListCmd = &cobra.Command{
	Use:   "list [keyword]",
	Short: "List all available templates",
	Long: fmt.Sprintf(`This command lists available Compute Instance templates. By default, returns "featured" templates.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&templateListOutput{}), ", ")),
	Aliases: gListAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var templateFilter string

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		community, err := cmd.Flags().GetBool("community")
		if err != nil {
			return err
		}

		mine, err := cmd.Flags().GetBool("mine")
		if err != nil {
			return err
		}

		if community {
			templateFilter = "community"
		} else if mine {
			templateFilter = "self"
		} else {
			templateFilter = "featured"
		}

		return printOutput(listTemplates(templateFilter, zone, args))
	},
}

func listTemplates(templateFilter, zone string, filters []string) (output.Outputter, error) {
	z, err := getZoneByNameOrID(zone)
	if err != nil {
		return nil, err
	}

	templates, err := findTemplates(z.ID, templateFilter, filters...)
	if err != nil {
		return nil, err
	}

	out := templateListOutput{}

	for _, template := range templates {
		out = append(out, templateListItemOutput{
			ID:           template.ID.String(),
			Name:         template.Name,
			DiskSize:     humanize.IBytes(uint64(template.Size)),
			CreationDate: template.Created,
			Zone:         template.ZoneName,
		})
	}

	return &out, nil
}
