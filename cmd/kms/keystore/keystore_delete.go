package keystore

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyStoreDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	KeyStore string      `cli-arg:"#" cli-usage:"ID"`
	Force    bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone     v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *keyStoreDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *keyStoreDeleteCmd) CmdShort() string     { return "Delete a key store" }
func (c *keyStoreDeleteCmd) CmdLong() string {
	return "This command deletes an external key store by ID."
}
func (c *keyStoreDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyStoreDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	id, err := parseKeyStoreID(c.KeyStore)
	if err != nil {
		return err
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete key store %q?", id)) {
			return nil
		}
	}

	if _, err := client.DeleteKeyStore(ctx, id); err != nil {
		return err
	}

	if !globalstate.Quiet {
		_, _ = fmt.Fprintln(os.Stdout, "Key store deleted.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &keyStoreDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
