package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var affinitygroupCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create Anti-Affinity Group",
	Long: fmt.Sprintf(`This command creates an Anti-Affinity Group.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&affinityGroupShowOutput{}), ", ")),
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		return output(createAffinityGroup(args[0], desc))
	},
}

func createAffinityGroup(name, desc string) (outputter, error) {
	resp, err := cs.RequestWithContext(gContext, &egoscale.CreateAffinityGroup{
		Name:        name,
		Description: desc,
		Type:        "host anti-affinity",
	})
	if err != nil {
		return nil, err
	}

	if !gQuiet {
		return showAffinityGroup(resp.(*egoscale.AffinityGroup))
	}

	return nil, nil
}

func init() {
	affinitygroupCreateCmd.Flags().StringP("description", "d", "", "affinity group description")
	affinitygroupCmd.AddCommand(affinitygroupCreateCmd)
}
