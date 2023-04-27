package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type privateNetworkLeaseOutput struct {
	Instance  string `json:"instance"`
	IPAddress string `json:"ip_address"`
}

type privateNetworkShowOutput struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Zone        string                      `json:"zone"`
	Type        string                      `json:"type"`
	StartIP     *string                     `json:"start_ip,omitempty"`
	EndIP       *string                     `json:"end_ip,omitempty"`
	Netmask     *string                     `json:"netmask,omitempty"`
	Leases      []privateNetworkLeaseOutput `json:"leases,omitempty"`
}

func (o *privateNetworkShowOutput) ToJSON() { output.JSON(o) }
func (o *privateNetworkShowOutput) ToText() { output.Text(o) }
func (o *privateNetworkShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Private Network"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"Type", o.Type})

	if o.Type == "managed" {
		t.Append([]string{"Start IP", *o.StartIP})
		t.Append([]string{"End IP", *o.EndIP})
		t.Append([]string{"Netmask", *o.Netmask})
		t.Append([]string{
			"Leases", func(leases []privateNetworkLeaseOutput) string {
				if len(leases) > 0 {
					buf := bytes.NewBuffer(nil)
					at := table.NewEmbeddedTable(buf)
					at.SetHeader([]string{" "})
					at.SetAlignment(tablewriter.ALIGN_LEFT)

					for _, lease := range leases {
						at.Append([]string{lease.Instance, lease.IPAddress})
					}
					at.Render()

					return buf.String()
				}
				return "-"
			}(o.Leases),
		})
	}
}

type privateNetworkShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"Private Network zone"`
}

func (c *privateNetworkShowCmd) cmdAliases() []string { return gShowAlias }

func (c *privateNetworkShowCmd) cmdShort() string {
	return "Show a Private Network details"
}

func (c *privateNetworkShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Private Network details.

Supported output template annotations for Private Network: %s

Supported output template annotations for Private Network leases: %s`,
		strings.Join(output.TemplateAnnotations(&privateNetworkShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&privateNetworkLeaseOutput{}), ", "))
}

func (c *privateNetworkShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	privateNetwork, err := globalstate.EgoscaleClient.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	out := privateNetworkShowOutput{
		ID:          *privateNetwork.ID,
		Zone:        c.Zone,
		Name:        *privateNetwork.Name,
		Description: utils.DefaultString(privateNetwork.Description, ""),
		Type:        "manual",
	}

	if privateNetwork.StartIP != nil {
		out.Type = "managed"

		startIP := privateNetwork.StartIP.String()
		out.StartIP = &startIP

		endIP := privateNetwork.EndIP.String()
		out.EndIP = &endIP

		netmask := privateNetwork.Netmask.String()
		out.Netmask = &netmask
	}

	if len(privateNetwork.Leases) > 0 {
		out.Leases = make([]privateNetworkLeaseOutput, 0)

		for _, lease := range privateNetwork.Leases {
			instance, err := globalstate.EgoscaleClient.GetInstance(ctx, c.Zone, *lease.InstanceID)
			if err != nil {
				return fmt.Errorf("unable to retrieve Compute instance %s: %w", *lease.InstanceID, err)
			}

			out.Leases = append(out.Leases, privateNetworkLeaseOutput{
				Instance:  *instance.Name,
				IPAddress: lease.IPAddress.String(),
			})
		}
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
