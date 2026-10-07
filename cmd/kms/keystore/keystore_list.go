package keystore

import (
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyStoreListOutput struct {
	v3.ListKeyStoresResponse
}

func (o *keyStoreListOutput) ToJSON() { output.JSON(o) }
func (o *keyStoreListOutput) ToText() { output.Text(o) }
func (o *keyStoreListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"ID",
		"NAME",
		"TYPE",
		"STATUS",
		"ENDPOINT",
	})

	for _, ks := range o.KeyStores {
		endpoint := ""
		if ks.Proxy != nil {
			endpoint = ks.Proxy.Endpoint
		}
		t.Append([]string{
			string(ks.ID),
			ks.Name,
			string(ks.Type),
			string(ks.Status),
			endpoint,
		})
	}
}

type keyStoreListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *keyStoreListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *keyStoreListCmd) CmdShort() string     { return "List key stores" }
func (c *keyStoreListCmd) CmdLong() string {
	return "This command lists the external key stores configured for the organization."
}
func (c *keyStoreListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyStoreListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	list, err := client.ListKeyStores(ctx)
	if err != nil {
		return err
	}

	return c.OutputFunc(&keyStoreListOutput{*list}, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &keyStoreListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
