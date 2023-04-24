package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var affinitygroupCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create an Anti-Affinity Group",
	Long: fmt.Sprintf(`This command creates an Anti-Affinity Group.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&affinityGroupShowOutput{}), ", ")),
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		return printOutput(createAffinityGroup(args[0], desc))
	},
}

func createAffinityGroup(name, desc string) (output.Outputter, error) {
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
	affinitygroupCreateCmd.Flags().StringP("description", "d", "", "Anti-Affinity Group description")
	affinitygroupCmd.AddCommand(affinitygroupCreateCmd)
}
