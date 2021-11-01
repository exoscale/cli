package cmd

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type nlbShowOutput struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	CreationDate string                 `json:"creation_date"`
	Zone         string                 `json:"zone"`
	IPAddress    string                 `json:"ip_address"`
	State        string                 `json:"state"`
	Services     []nlbServiceShowOutput `json:"services"`
	Labels       map[string]string      `json:"labels"`
}

func (o *nlbShowOutput) toJSON() { outputJSON(o) }
func (o *nlbShowOutput) toText() { outputText(o) }
func (o *nlbShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Network Load Balancer"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Zone", o.Zone})
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbShowCmd) cmdAliases() []string { return gShowAlias }

func (c *nlbShowCmd) cmdShort() string { return "Show a Network Load Balancer details" }

func (c *nlbShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Network Load Balancer details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&nlbShowOutput{}), ", "))
}

func (c *nlbShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	nlb, err := cs.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	svcOut := make([]nlbServiceShowOutput, 0)
	for _, svc := range nlb.Services {
		svcOut = append(svcOut, nlbServiceShowOutput{
			ID:   *svc.ID,
			Name: *svc.Name,
		})
	}

	out := nlbShowOutput{
		ID:           *nlb.ID,
		Name:         *nlb.Name,
		Description:  defaultString(nlb.Description, ""),
		CreationDate: nlb.CreatedAt.String(),
		Zone:         c.Zone,
		IPAddress:    nlb.IPAddress.String(),
		State:        *nlb.State,
		Services:     svcOut,
		Labels: func() (v map[string]string) {
			if nlb.Labels != nil {
				v = *nlb.Labels
			}
			return
		}(),
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedNLBCmd, &nlbShowCmd{}))
}
