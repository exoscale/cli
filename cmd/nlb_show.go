package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exoapi "github.com/exoscale/egoscale/v2/api"

	"github.com/exoscale/cli/table"
)

type nlbShowOutput struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	CreationDate string                 `json:"created_at"`
	Zone         string                 `json:"zone"`
	IPAddress    string                 `json:"ip_address"`
	State        string                 `json:"state"`
	Services     []nlbServiceShowOutput `json:"services"`
}

func (o *nlbShowOutput) toJSON() { outputJSON(o) }
func (o *nlbShowOutput) toText() { outputText(o) }
func (o *nlbShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"Network Load Balancer"})
	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"IP Address", o.IPAddress})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Creation Date", o.CreationDate})
	t.Append([]string{"Services", func() string {
		if len(o.Services) > 0 {
			return strings.Join(
				func() []string {
					services := make([]string, len(o.Services))
					for i := range o.Services {
						services[i] = fmt.Sprintf("%s | %s",
							o.Services[i].ID,
							o.Services[i].Name)
					}
					return services
				}(),
				"\n")
		}
		return "n/a"
	}()})
	t.Append([]string{"State", o.State})
}

var nlbShowCmd = &cobra.Command{
	Use:   "show <name | ID>",
	Short: "Show a Network Load Balancer details",
	Long: fmt.Sprintf(`This command shows a Network Load Balancer details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&nlbShowOutput{}), ", ")),
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

		return output(showNLB(zone, args[0]))
	},
}

func showNLB(zone, ref string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	nlb, err := lookupNLB(ctx, zone, ref)
	if err != nil {
		return nil, err
	}

	svcOut := make([]nlbServiceShowOutput, 0)
	for _, svc := range nlb.Services {
		svcOut = append(svcOut, nlbServiceShowOutput{
			ID:   svc.ID,
			Name: svc.Name,
		})
	}

	out := nlbShowOutput{
		ID:           nlb.ID,
		Name:         nlb.Name,
		Description:  nlb.Description,
		CreationDate: nlb.CreatedAt.String(),
		Zone:         zone,
		IPAddress:    nlb.IPAddress.String(),
		State:        nlb.State,
		Services:     svcOut,
	}

	return &out, nil
}

func init() {
	nlbShowCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbCmd.AddCommand(nlbShowCmd)
}
