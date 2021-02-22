package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksClusterVersionsItemOutput struct {
	Version string `json:"version"`
}

type sksClusterVersionsOutput []sksClusterVersionsItemOutput

func (o *sksClusterVersionsOutput) toJSON()  { outputJSON(o) }
func (o *sksClusterVersionsOutput) toText()  { outputText(o) }
func (o *sksClusterVersionsOutput) toTable() { outputTable(o) }

var sksVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List SKS cluster versions",
	Long: fmt.Sprintf(`This command lists supported SKS cluster versions.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksClusterVersionsItemOutput{}), ", ")),
	Aliases: gListAlias,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone = strings.ToLower(zone)

		return output(listSKSVersions(zone))
	},
}

func listSKSVersions(zone string) (outputter, error) {
	out := make(sksClusterVersionsOutput, 0)

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	versions, err := cs.ListSKSClusterVersions(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range versions {
		out = append(out, sksClusterVersionsItemOutput{Version: v})
	}

	return &out, nil
}

func init() {
	sksVersionsCmd.Flags().StringP("zone", "z", "", "Zone to filter results to")
	sksCmd.AddCommand(sksVersionsCmd)
}
