package load_balancer

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbShowOutput struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	CreationDate string                 `json:"creation_date"`
	Zone         v3.ZoneName            `json:"zone"`
	IPAddress    string                 `json:"ip_address"`
	State        string                 `json:"state"`
	Services     []nlbServiceShowOutput `json:"services"`
	Labels       map[string]string      `json:"labels"`
}

func (o *nlbShowOutput) ToJSON() { output.JSON(o) }
func (o *nlbShowOutput) ToText() { output.Text(o) }
func (o *nlbShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Network Load Balancer"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Zone", string(o.Zone)})
	t.Append([]string{"IP Address", o.IPAddress})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Creation Date", o.CreationDate})
	t.Append([]string{"State", o.State})

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

	t.Append([]string{"Labels", func() string {
		sortedKeys := func() []string {
			keys := make([]string, 0)
			for k := range o.Labels {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return keys
		}()

		buf := bytes.NewBuffer(nil)
		at := table.NewEmbeddedTable(buf)
		at.SetHeader([]string{" "})
		for _, k := range sortedKeys {
			at.Append([]string{k, o.Labels[k]})
		}
		at.Render()

		return buf.String()
	}()})
}

type nlbShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *nlbShowCmd) CmdShort() string { return "Show a Network Load Balancer details" }

func (c *nlbShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Network Load Balancer details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbShowOutput{}), ", "))
}

func (c *nlbShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	nlbs, err := client.ListLoadBalancers(ctx)
	if err != nil {
		return err
	}

	nlb, err := nlbs.FindLoadBalancer(c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	svcOut := make([]nlbServiceShowOutput, 0)
	for _, svc := range nlb.Services {
		svcOut = append(svcOut, nlbServiceShowOutput{
			ID:   string(svc.ID),
			Name: svc.Name,
		})
	}

	out := nlbShowOutput{
		ID:           nlb.ID.String(),
		Name:         nlb.Name,
		Description:  nlb.Description,
		CreationDate: nlb.CreatedAT.String(),
		Zone:         c.Zone,
		IPAddress:    nlb.IP.String(),
		State:        string(nlb.State),
		Services:     svcOut,
		Labels:       nlb.Labels,
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
