package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
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
	Type        string                      `json:"type"`
	StartIP     *string                     `json:"start_ip,omitempty"`
	EndIP       *string                     `json:"end_ip,omitempty"`
	Netmask     *string                     `json:"netmask,omitempty"`
	Leases      []privateNetworkLeaseOutput `json:"leases,omitempty"`
}

func (o *privateNetworkShowOutput) toJSON() { outputJSON(o) }
func (o *privateNetworkShowOutput) toText() { outputText(o) }
func (o *privateNetworkShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Private Network"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
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
		strings.Join(outputterTemplateAnnotations(&privateNetworkShowOutput{}), ", "),
		strings.Join(outputterTemplateAnnotations(&privateNetworkLeaseOutput{}), ", "))
}

func (c *privateNetworkShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return output(showPrivateNetwork(c.Zone, c.PrivateNetwork))
}

func showPrivateNetwork(zone, x string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	privateNetwork, err := cs.FindPrivateNetwork(ctx, zone, x)
	if err != nil {
		return nil, err
	}

	out := privateNetworkShowOutput{
		ID:          *privateNetwork.ID,
		Name:        *privateNetwork.Name,
		Description: defaultString(privateNetwork.Description, ""),
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
			instance, err := cs.GetInstance(ctx, zone, *lease.InstanceID)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve Compute instance %s: %v", *lease.InstanceID, err)
			}

			out.Leases = append(out.Leases, privateNetworkLeaseOutput{
				Instance:  *instance.Name,
				IPAddress: lease.IPAddress.String(),
			})
		}
	}

	return &out, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
