package cmd

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type privateNetworkLeaseOutput struct {
	Instance  string `json:"instance"`
	IPAddress string `json:"ip_address"`
}

type privateNetworkOptionsOutput struct {
	Routers      []net.IP `json:"routers"`
	DNSServers   []net.IP `json:"dns-servers"`
	NTPServers   []net.IP `json:"ntp-servers"`
	DomainSearch []string `json:"domain-search"`
}

type privateNetworkShowOutput struct {
	ID          v3.UUID                     `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Zone        v3.ZoneName                 `json:"zone"`
	Type        string                      `json:"type"`
	StartIP     *string                     `json:"start_ip,omitempty"`
	EndIP       *string                     `json:"end_ip,omitempty"`
	Netmask     *string                     `json:"netmask,omitempty"`
	Leases      []privateNetworkLeaseOutput `json:"leases,omitempty"`
	Options     privateNetworkOptionsOutput `json:"options"`
}

func (o *privateNetworkShowOutput) ToJSON() { output.JSON(o) }
func (o *privateNetworkShowOutput) ToText() { output.Text(o) }
func (o *privateNetworkShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Private Network"})
	defer t.Render()

	t.Append([]string{"ID", o.ID.String()})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", string(o.Zone)})
	t.Append([]string{"Type", o.Type})

	if o.Type == "managed" {
		t.Append([]string{"Start IP", *o.StartIP})
		t.Append([]string{"End IP", *o.EndIP})
		t.Append([]string{"Netmask", *o.Netmask})

		t.Append([]string{
			"Leases", formatLeases(o.Leases),
		})
	}
	t.Append([]string{
		"Options", formatOptions(o.Options),
	})
}

type privateNetworkShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"Private Network zone"`
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
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListPrivateNetworks(ctx)
	if err != nil {
		return err
	}

	pn, err := resp.FindPrivateNetwork(c.PrivateNetwork)
	if err != nil {
		return err
	}

	privateNetwork, err := client.GetPrivateNetwork(ctx, pn.ID)
	if err != nil {
		return err
	}

	out := privateNetworkShowOutput{
		ID:          privateNetwork.ID,
		Zone:        c.Zone,
		Name:        privateNetwork.Name,
		Description: privateNetwork.Description,
		Type:        "manual",
		Options: func() privateNetworkOptionsOutput {
			if privateNetwork.Options == nil {
				return privateNetworkOptionsOutput{}
			}
			return privateNetworkOptionsOutput{
				Routers:      privateNetwork.Options.Routers,
				DNSServers:   privateNetwork.Options.DNSServers,
				NTPServers:   privateNetwork.Options.NtpServers,
				DomainSearch: privateNetwork.Options.DomainSearch,
			}
		}(),
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
			instance, err := client.GetInstance(ctx, lease.InstanceID)
			if err != nil {
				return fmt.Errorf("unable to retrieve Compute instance %s: %w", lease.InstanceID, err)
			}

			out.Leases = append(out.Leases, privateNetworkLeaseOutput{
				Instance:  instance.Name,
				IPAddress: lease.IP.String(),
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

func ipSliceToStringSlice(ips []net.IP) []string {
	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}
	return result
}

func formatLeases(leases []privateNetworkLeaseOutput) string {
	if len(leases) == 0 {
		return "-"
	}

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

func formatOptions(opts privateNetworkOptionsOutput) string {
	hasOptions := len(opts.Routers) > 0 || len(opts.DNSServers) > 0 ||
		len(opts.NTPServers) > 0 || len(opts.DomainSearch) > 0

	if !hasOptions {
		return "-"
	}

	buf := bytes.NewBuffer(nil)
	at := table.NewEmbeddedTable(buf)
	at.SetHeader([]string{" "})
	at.SetAlignment(tablewriter.ALIGN_LEFT)

	if len(opts.Routers) > 0 {
		at.Append([]string{"Routers", strings.Join(ipSliceToStringSlice(opts.Routers), ", ")})
	}
	if len(opts.DNSServers) > 0 {
		at.Append([]string{"DNS Servers", strings.Join(ipSliceToStringSlice(opts.DNSServers), ", ")})
	}
	if len(opts.NTPServers) > 0 {
		at.Append([]string{"NTP Servers", strings.Join(ipSliceToStringSlice(opts.NTPServers), ", ")})
	}
	if len(opts.DomainSearch) > 0 {
		at.Append([]string{"Domain Search", strings.Join(opts.DomainSearch, ", ")})
	}

	at.Render()
	return buf.String()
}
