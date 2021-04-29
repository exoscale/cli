package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exoapi "github.com/exoscale/egoscale/v2/api"
)

type deployTargetShowOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Zone        string `json:"zone"`
}

func (o *deployTargetShowOutput) toJSON()  { outputJSON(o) }
func (o *deployTargetShowOutput) toText()  { outputText(o) }
func (o *deployTargetShowOutput) toTable() { outputTable(o) }

var deployTargetShowCmd = &cobra.Command{
	Use:   "show NAME|ID",
	Short: "Show a Deploy Target details",
	Long: fmt.Sprintf(`This command shows a Deploy Target details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&deployTargetShowOutput{}), ", ")),
	Aliases: gShowAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		return output(showDeployTarget(zone, args[0]))
	},
}

func showDeployTarget(zone, c string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	dt, err := lookupDeployTarget(ctx, zone, c)
	if err != nil {
		return nil, err
	}

	return &deployTargetShowOutput{
		ID:          dt.ID,
		Name:        dt.Name,
		Description: dt.Description,
		Type:        dt.Type,
		Zone:        zone,
	}, nil
}

func init() {
	deployTargetShowCmd.Flags().StringP("zone", "z", "", "Deploy Target zone")
	deployTargetCmd.AddCommand(deployTargetShowCmd)
}
