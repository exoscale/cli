package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/spf13/cobra"
)

type runstatusPageListItemOutput struct {
	ID        int    `json:"id" output:"-"`
	Name      string `json:"name"`
	PublicURL string `json:"public_url"`
}

type runstatusPageListOutput []runstatusPageListItemOutput

func (o *runstatusPageListOutput) toJSON()  { output.JSON(o) }
func (o *runstatusPageListOutput) toText()  { output.Text(o) }
func (o *runstatusPageListOutput) toTable() { output.Table(o) }

func init() {
	runstatusCmd.AddCommand(&cobra.Command{
		Use:   "list [FILTER]...",
		Short: "List runstat.us pages",
		Long: fmt.Sprintf(`This command lists existing runstat.us pages.
Optional patterns can be provided to filter results by name.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&runstatusPageListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listRunstatusPages(args))
		},
	})
}

func listRunstatusPages(filters []string) (outputter, error) {
	pages, err := csRunstatus.ListRunstatusPages(gContext)
	if err != nil {
		return nil, err
	}

	out := runstatusPageListOutput{}

	for _, p := range pages {
		keep := true
		if len(filters) > 0 {
			keep = false
			s := strings.ToLower(p.Subdomain)

			for _, filter := range filters {
				substr := strings.ToLower(filter)
				if strings.Contains(s, substr) {
					keep = true
					break
				}
			}
		}

		if !keep {
			continue
		}

		out = append(out, runstatusPageListItemOutput{
			ID:        p.ID,
			Name:      p.Subdomain,
			PublicURL: p.PublicURL,
		})
	}

	return &out, nil
}
