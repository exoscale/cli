package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceTemplateShowOutput struct {
	ID                           string `json:"id"`
	Zone                         string `json:"zone"`
	Name                         string `json:"name"`
	Description                  string `json:"description"`
	Family                       string `json:"family"`
	CreationDate                 string `json:"creation_date"`
	Visibility                   string `json:"visibility"`
	Size                         int64  `json:"size"`
	Version                      string `json:"version"`
	Build                        string `json:"build"`
	Maintainer                   string `json:"maintainer"`
	DefaultUser                  string `json:"default_user"`
	SSHKeyEnabled                bool   `json:"ssh_key_enabled"`
	PasswordEnabled              bool   `json:"password_enabled"`
	BootMode                     string `json:"boot_mode"`
	Checksum                     string `json:"checksum"`
	AppConsistentSnapshotEnabled bool   `json:"application_consistent_snapshot_enabled"`
}

func (o *instanceTemplateShowOutput) ToJSON() { output.JSON(o) }
func (o *instanceTemplateShowOutput) ToText() { output.Text(o) }
func (o *instanceTemplateShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Template"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Family", o.Family})
	t.Append([]string{"Creation Date", o.CreationDate})
	t.Append([]string{"Visibility", o.Visibility})
	t.Append([]string{"Size", humanize.IBytes(uint64(o.Size))})
	t.Append([]string{"Version", o.Version})
	t.Append([]string{"Build", o.Build})
	t.Append([]string{"Maintainer", o.Maintainer})
	t.Append([]string{"Default User", o.DefaultUser})
	t.Append([]string{"SSH key enabled", fmt.Sprint(o.SSHKeyEnabled)})
	t.Append([]string{"Password enabled", fmt.Sprint(o.PasswordEnabled)})
	t.Append([]string{"Boot Mode", o.BootMode})
	t.Append([]string{"Checksum", o.Checksum})
	t.Append([]string{"Application Consistent Snapshot Enabled", fmt.Sprint(o.AppConsistentSnapshotEnabled)})
}

type instanceTemplateShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Template string `cli-arg:"#" cli-usage:"[FAMILY.]SIZE"`

	Visibility string `cli-short:"v" cli-usage:"template visibility (public|private)"`
	Zone       string `cli-short:"z" cli-usage:"zone to filter results to (default: current account's default zone)"`
}

func (c *instanceTemplateShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *instanceTemplateShowCmd) CmdShort() string {
	return "Show a Compute instance template details"
}

func (c *instanceTemplateShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance template details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceTemplateShowOutput{}), ", "))
}

func (c *instanceTemplateShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	templates, err := client.ListTemplates(ctx, v3.ListTemplatesWithVisibility(v3.ListTemplatesVisibility(c.Visibility)))
	if err != nil {
		return err
	}

	template, err := templates.FindTemplate(c.Template)
	if err != nil {
		return err
	}

	return c.OutputFunc(&instanceTemplateShowOutput{
		ID:                           template.ID.String(),
		Zone:                         c.Zone,
		Family:                       template.Family,
		Name:                         template.Name,
		Description:                  template.Description,
		CreationDate:                 template.CreatedAT.String(),
		Visibility:                   string(template.Visibility),
		Size:                         template.Size,
		Version:                      template.Version,
		Build:                        template.Build,
		Maintainer:                   template.Maintainer,
		Checksum:                     template.Checksum,
		DefaultUser:                  template.DefaultUser,
		SSHKeyEnabled:                utils.DefaultBool(template.SSHKeyEnabled, false),
		PasswordEnabled:              utils.DefaultBool(template.PasswordEnabled, false),
		BootMode:                     string(template.BootMode),
		AppConsistentSnapshotEnabled: utils.DefaultBool(template.ApplicationConsistentSnapshotEnabled, false),
	}, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceTemplateCmd, &instanceTemplateShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		Visibility: exocmd.DefaultTemplateVisibility,
	}))
}
