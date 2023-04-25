package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
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
func (o *runstatusServiceShowOutput) ToJSON()      { output.JSON(o) }
func (o *runstatusServiceShowOutput) ToText()      { output.Text(o) }
func (o *runstatusServiceShowOutput) ToTable()     { output.Table(o) }

func init() {
	runstatusServiceCmd.AddCommand(&cobra.Command{
		Use:   "show [PAGE] SERVICE-NAME",
		Short: "Show a service details",
		Long: fmt.Sprintf(`This command shows a runstat.us page details.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&runstatusServiceShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			page := account.CurrentAccount.DefaultRunstatusPage
			service := args[0]

			if account.CurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
				return fmt.Errorf("No default runstat.us page is set.\n"+
					"Please specify a page in parameter or add it to your configuration file %s",
					gConfigFilePath)
			}

			if len(args) > 1 {
				page = args[0]
				service = args[1]
			}

			return printOutput(runstatusShowService(page, service))
		},
	})
}

func runstatusShowService(p, s string) (output.Outputter, error) {
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
