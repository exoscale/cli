package key

import (
	"os"
	"strconv"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyListOutput struct {
	v3.ListKmsKeysResponse
}

func (o *keyListOutput) ToJSON() { output.JSON(o) }
func (o *keyListOutput) ToText() { output.Text(o) }
func (o *keyListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"ID",
		"NAME",
		"ORIGINZONE",
		"STATUS",
		"MULTIZONE",
		"REPLICAS",
	})

	for _, key := range o.KmsKeys {
		t.Append([]string{
			string(key.ID),
			key.Name,
			string(key.OriginZone),
			string(key.Status),
			strconv.FormatBool(*key.MultiZone),
			strings.Join(key.Replicas, ", "),
		})
	}
}

type keyListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	IgnoreReplica bool        `cli-short:"i" cli-flag:"ignore-replica" cli-usage:"filter out replicas"`
	Status        string      `cli-short:"s" cli-flag:"status" cli-usage:"filter by key status [enabled|disabled|pending-deletion]"`
	Zone          v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *keyListCmd) CmdShort() string {
	return "List KMS Keys details for an organization in a given zone."
}

func (c *keyListCmd) CmdLong() string {
	return "List KMS Keys details for an organization in a given zone."
}

func (c *keyListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	keys, err := client.ListKmsKeys(ctx)
	if err != nil {
		return err
	}

	filtered := make([]v3.ListKmsKeysResponseEntry, 0, len(keys.KmsKeys))
	for _, key := range keys.KmsKeys {
		if c.IgnoreReplica && key.OriginZone != string(c.Zone) {
			continue
		}
		if c.Status != "" && string(key.Status) != c.Status {
			continue
		}
		filtered = append(filtered, key)
	}

	out := keyListOutput{v3.ListKmsKeysResponse{KmsKeys: filtered}}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
