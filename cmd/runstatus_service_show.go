package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type runstatusServiceShowOutput struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

func (o *runstatusServiceShowOutput) Type() string { return "Service" }
func (o *runstatusServiceShowOutput) toJSON()      { output.JSON(o) }
func (o *runstatusServiceShowOutput) toText()      { output.Text(o) }
func (o *runstatusServiceShowOutput) toTable()     { output.Table(o) }

func init() {
	runstatusServiceCmd.AddCommand(&cobra.Command{
		Use:   "show [PAGE] SERVICE-NAME",
		Short: "Show a service details",
		Long: fmt.Sprintf(`This command shows a runstat.us page details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&runstatusServiceShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			page := gCurrentAccount.DefaultRunstatusPage
			service := args[0]

			if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
				return fmt.Errorf("No default runstat.us page is set.\n"+
					"Please specify a page in parameter or add it to your configuration file %s",
					gConfigFilePath)
			}

			if len(args) > 1 {
				page = args[0]
				service = args[1]
			}

			return output(runstatusShowService(page, service))
		},
	})
}

func runstatusShowService(p, s string) (outputter, error) {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: p})
	if err != nil {
		return nil, err
	}

	service, err := getServiceByName(*page, s)
	if err != nil {
		return nil, err
	}

	out := runstatusServiceShowOutput{
		ID:    service.ID,
		Name:  service.Name,
		State: service.State,
	}

	return &out, nil
}
