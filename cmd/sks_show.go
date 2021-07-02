package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksShowOutput struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	CreationDate string                  `json:"creation_date"`
	Zone         string                  `json:"zone"`
	Endpoint     string                  `json:"endpoint"`
	Version      string                  `json:"version"`
	ServiceLevel string                  `json:"service_level"`
	CNI          string                  `json:"cni"`
	AddOns       []string                `json:"addons"`
	State        string                  `json:"state"`
	Labels       map[string]string       `json:"labels"`
	Nodepools    []sksNodepoolShowOutput `json:"nodepools"`
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
	t.Append([]string{"Labels", func() string {
		if len(o.Labels) > 0 {
			return strings.Join(
				func() []string {
					labels := make([]string, 0)
					for k, v := range o.Labels {
						labels = append(labels, fmt.Sprintf("%s:%s", k, v))
					}
					return labels
				}(),
				"\n")
		}
		return "n/a"
	}()})
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

type sksShowCmd struct {
	_ bool `cli-cmd:"show"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksShowCmd) cmdAliases() []string { return gShowAlias }

func (c *sksShowCmd) cmdShort() string { return "Show an SKS cluster details" }

func (c *sksShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows an SKS cluster details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksShowOutput{}), ", "))
}

func (c *sksShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return output(showSKSCluster(c.Zone, c.Cluster))
}

func showSKSCluster(zone, c string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	cluster, err := cs.FindSKSCluster(ctx, zone, c)
	if err != nil {
		return nil, err
	}

	sksNodepools := make([]sksNodepoolShowOutput, 0)
	for _, np := range cluster.Nodepools {
		sksNodepools = append(sksNodepools, sksNodepoolShowOutput{
			ID:   *np.ID,
			Name: *np.Name,
		})
	}

	out := sksShowOutput{
		AddOns: func() (v []string) {
			if cluster.AddOns != nil {
				v = *cluster.AddOns
			}
			return
		}(),
		CNI:          defaultString(cluster.CNI, "-"),
		CreationDate: cluster.CreatedAt.String(),
		Description:  defaultString(cluster.Description, ""),
		Endpoint:     *cluster.Endpoint,
		ID:           *cluster.ID,
		Labels: func() (v map[string]string) {
			if cluster.Labels != nil {
				v = *cluster.Labels
			}
			return
		}(),
		Name:         *cluster.Name,
		Nodepools:    sksNodepools,
		ServiceLevel: *cluster.ServiceLevel,
		State:        *cluster.State,
		Version:      *cluster.Version,
		Zone:         zone,
	}

	return &out, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksShowCmd{}))
}
