package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasTypeListItemOutput struct {
	Name              string   `json:"name"`
	AvailableVersions []string `json:"available_versions"`
	DefaultVersion    string   `json:"default_version"`
}

type dbaasTypeListOutput []dbaasTypeListItemOutput

func (o *dbaasTypeListOutput) ToJSON() { output.JSON(o) }
func (o *dbaasTypeListOutput) ToText() { output.Text(o) }
func (o *dbaasTypeListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Name", "Available Versions", "Default Version"})
	defer t.Render()

	for _, dbType := range *o {
		t.Append([]string{
			dbType.Name,
			strings.Join(dbType.AvailableVersions, ", "),
			dbType.DefaultVersion,
		})
	}
}

type dbaasTypeListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *dbaasTypeListCmd) cmdAliases() []string { return nil }

func (c *dbaasTypeListCmd) cmdShort() string { return "List Database Service types" }

func (c *dbaasTypeListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Database Service types.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&dbaasTypeListItemOutput{}), ", "))
}

func (c *dbaasTypeListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasTypeListCmd) cmdRun(_ *cobra.Command, _ []string) error {

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	dbTypes, err := client.ListDBAASServiceTypes(ctx)
	if err != nil {
		return err
	}

	out := make(dbaasTypeListOutput, 0)

	for _, t := range dbTypes.DBAASServiceTypes {
		out = append(out, dbaasTypeListItemOutput{
			Name:           string(t.Name),
			DefaultVersion: utils.DefaultString(&t.DefaultVersion, "-"),
			AvailableVersions: func() (v []string) {
				if t.AvailableVersions != nil {
					v = t.AvailableVersions
				}
				return
			}(),
		})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasTypeCmd, &dbaasTypeListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
