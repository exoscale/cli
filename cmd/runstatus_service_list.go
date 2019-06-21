package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type runstatusServiceListItemOutput struct {
	ID      int    `page:"id" output:"-"`
	Service string `json:"service"`
	State   string `json:"state"`
	Page    string `page:"page"`
}

type runstatusServiceListOutput []runstatusServiceListItemOutput

func (o *runstatusServiceListOutput) toJSON() { outputJSON(o) }

func (o *runstatusServiceListOutput) toText() { outputText(o) }

func (o *runstatusServiceListOutput) toTable() {
	for i := range *o {
		(*o)[i].State = strings.ToUpper(strings.Replace((*o)[i].State, "_", " ", -1))
	}

	outputTable(o)
}

func init() {
	runstatusServiceCmd.AddCommand(&cobra.Command{
		Use:   "list [page name ...]",
		Short: "List services",
		Long: fmt.Sprintf(`This command lists existing runstat.us services.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&runstatusServiceListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(runstatusListServices(args))
		},
	})
}

func runstatusListServices(pageNames []string) (outputter, error) {
	pages, err := getRunstatusPages(pageNames)
	if err != nil {
		return nil, err
	}

	out := runstatusServiceListOutput{}

	for _, page := range pages {
		services, err := csRunstatus.ListRunstatusServices(gContext, page)
		if err != nil {
			return nil, err
		}

		for _, service := range services {
			out = append(out, runstatusServiceListItemOutput{
				ID:      service.ID,
				Service: service.Name,
				State:   service.State,
				Page:    page.Subdomain,
			})
		}
	}

	return &out, nil
}
