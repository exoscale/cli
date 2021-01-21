package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	apiv2 "github.com/exoscale/egoscale/api/v2"

	"github.com/exoscale/cli/table"
)

type sksShowOutput struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	CreationDate string                 `json:"creation_date"`
	Zone         string                 `json:"zone"`
	Endpoint     string                 `json:"endpoint"`
	Version      string                 `json:"version"`
	ServiceLevel string                 `json:"service_level"`
	CNI          string                 `json:"cni"`
	AddOns       []string               `json:"addons"`
	State        string                 `json:"state"`
	Nodepools    []nlbServiceShowOutput `json:"nodepools"`
}

func (o *sksShowOutput) toJSON() { outputJSON(o) }
func (o *sksShowOutput) toText() { outputText(o) }
func (o *sksShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"SKS Cluster"})
	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"Creation Date", o.CreationDate})
	t.Append([]string{"Endpoint", o.Endpoint})
	t.Append([]string{"Version", o.Version})
	t.Append([]string{"Service Level", o.ServiceLevel})
	t.Append([]string{"CNI", o.CNI})
	t.Append([]string{"Add-Ons", strings.Join(o.AddOns, "\n")})
	t.Append([]string{"State", o.State})
	t.Append([]string{"Nodepools", func() string {
		if len(o.Nodepools) > 0 {
			return strings.Join(
				func() []string {
					services := make([]string, len(o.Nodepools))
					for i := range o.Nodepools {
						services[i] = fmt.Sprintf("%s | %s",
							o.Nodepools[i].ID,
							o.Nodepools[i].Name)
					}
					return services
				}(),
				"\n")
		}
		return "n/a"
	}()})
}

var sksShowCmd = &cobra.Command{
	Use:   "show <name | ID>",
	Short: "Show a SKS cluster details",
	Long: fmt.Sprintf(`This command shows a SKS cluster details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksShowOutput{}), ", ")),
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

		return output(showSKSCluster(zone, args[0]))
	},
}

func showSKSCluster(zone, c string) (outputter, error) {
	ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
	cluster, err := lookupSKSCluster(ctx, zone, c)
	if err != nil {
		return nil, err
	}

	sksNodepools := make([]nlbServiceShowOutput, 0)
	for _, np := range cluster.Nodepools {
		sksNodepools = append(sksNodepools, nlbServiceShowOutput{
			ID:   np.ID,
			Name: np.Name,
		})
	}

	out := sksShowOutput{
		ID:           cluster.ID,
		Name:         cluster.Name,
		Description:  cluster.Description,
		CreationDate: cluster.CreatedAt.String(),
		Version:      cluster.Version,
		ServiceLevel: cluster.Level,
		CNI:          cluster.CNI,
		AddOns:       cluster.AddOns,
		Zone:         zone,
		Endpoint:     cluster.Endpoint,
		State:        cluster.State,
		Nodepools:    sksNodepools,
	}

	return &out, nil
}

func init() {
	sksShowCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksCmd.AddCommand(sksShowCmd)
}
