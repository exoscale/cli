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

	KeyStores []string    `cli-arg:"#" cli-usage:"NAME|ID..."`
	Force     bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone      v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *keyStoreDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }
func (c *keyStoreDeleteCmd) CmdShort() string     { return "Delete key stores" }
func (c *keyStoreDeleteCmd) CmdLong() string {
	return "This command deletes one or more external key stores by name or ID."
}
func (c *keyStoreDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyStoreDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	list, err := client.ListKeyStores(ctx)
	if err != nil {
		return err
	}

	toDelete := []v3.UUID{}
	for _, ksStr := range c.KeyStores {
		entry, err := list.FindListKeyStoresResponseEntry(ksStr)
		if err != nil {
			if !c.Force {
				return err
			}
			fmt.Fprintf(os.Stderr, "warning: %s not found.\n", ksStr)
			continue
		}

		if !c.Force {
			if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete key store %q?", ksStr)) {
				return nil
			}
		}

		toDelete = append(toDelete, entry.ID)
	}

	var fns []func() error
	for _, id := range toDelete {
		fns = append(fns, func() error {
			_, err := client.DeleteKeyStore(ctx, id)
			return err
		})
	}

	if err := utils.DecorateAsyncOperations("Deleting key store(s)...", fns...); err != nil {
		return err
	}

	if !globalstate.Quiet {
		_, _ = fmt.Fprintln(os.Stdout, "Key store(s) deleted.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &keyStoreDeleteCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
