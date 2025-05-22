package sks

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksShowOutput struct {
	ID              v3.UUID                 `json:"id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	CreationDate    string                  `json:"creation_date"`
	AutoUpgrade     bool                    `json:"auto_upgrade"`
	EnableKubeProxy bool                    `json:"enable_kube_proxy"`
	Zone            v3.ZoneName             `json:"zone"`
	Endpoint        string                  `json:"endpoint"`
	Version         string                  `json:"version"`
	ServiceLevel    string                  `json:"service_level"`
	CNI             string                  `json:"cni"`
	AddOns          []string                `json:"addons"`
	FeatureGates    []string                `json:"feature_gates"`
	State           string                  `json:"state"`
	Labels          map[string]string       `json:"labels"`
	Nodepools       []sksNodepoolShowOutput `json:"nodepools"`
}

func (o *sksShowOutput) ToJSON() { output.JSON(o) }
func (o *sksShowOutput) ToText() { output.Text(o) }
func (o *sksShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"SKS Cluster"})
	defer t.Render()

	t.Append([]string{"ID", o.ID.String()})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", string(o.Zone)})
	t.Append([]string{"Creation Date", o.CreationDate})
	t.Append([]string{"Auto-upgrade", fmt.Sprint(o.AutoUpgrade)})
	t.Append([]string{"Enable kube-proxy", fmt.Sprint(o.EnableKubeProxy)})
	t.Append([]string{"Endpoint", o.Endpoint})
	t.Append([]string{"Version", o.Version})
	t.Append([]string{"Service Level", o.ServiceLevel})
	t.Append([]string{"CNI", o.CNI})
	t.Append([]string{"Add-Ons", strings.Join(o.AddOns, "\n")})
	t.Append([]string{"Feature Gates", strings.Join(o.FeatureGates, "\n")})
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
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *sksShowCmd) CmdShort() string { return "Show an SKS cluster details" }

func (c *sksShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows an SKS cluster details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksShowOutput{}), ", "))
}

func (c *sksShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	clusters, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := clusters.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	sksNodepools := make([]sksNodepoolShowOutput, 0)
	for _, np := range cluster.Nodepools {
		sksNodepools = append(sksNodepools, sksNodepoolShowOutput{
			ID:   np.ID,
			Name: np.Name,
		})
	}

	return c.OutputFunc(
		&sksShowOutput{
			AddOns: func() (v []string) {
				if cluster.Addons != nil {
					v = cluster.Addons
				}
				return
			}(),
			CNI:             string(cluster.Cni),
			CreationDate:    cluster.CreatedAT.String(),
			AutoUpgrade:     *cluster.AutoUpgrade,
			EnableKubeProxy: *cluster.EnableKubeProxy,
			Description:     cluster.Description,
			Endpoint:        cluster.Endpoint,
			FeatureGates: func() (v []string) {
				if cluster.FeatureGates != nil {
					v = cluster.FeatureGates
				}
				return
			}(),
			ID: cluster.ID,
			Labels: func() (v map[string]string) {
				if cluster.Labels != nil {
					v = cluster.Labels
				}
				return
			}(),
			Name:         cluster.Name,
			Nodepools:    sksNodepools,
			ServiceLevel: string(cluster.Level),
			State:        string(cluster.State),
			Version:      cluster.Version,
			Zone:         c.Zone,
		},
		nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
