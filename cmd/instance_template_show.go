package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceTemplateShowOutput struct {
	ID              string `json:"id"`
	Zone            string `json:"zone"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Family          string `json:"family"`
	CreationDate    string `json:"creation_date"`
	Visibility      string `json:"visibility"`
	Size            int64  `json:"size"`
	Version         string `json:"version"`
	Build           string `json:"build"`
	Maintainer      string `json:"maintainer"`
	DefaultUser     string `json:"default_user"`
	SSHKeyEnabled   bool   `json:"ssh_key_enabled"`
	PasswordEnabled bool   `json:"password_enabled"`
	BootMode        string `json:"boot_mode"`
	Checksum        string `json:"checksum"`
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
}

type instanceTemplateShowCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Template string `cli-arg:"#" cli-usage:"[FAMILY.]SIZE"`

	Visibility string `cli-short:"v" cli-usage:"template visibility (public|private)"`
	Zone       string `cli-short:"z" cli-usage:"zone to filter results to (default: current account's default zone)"`
}

func (c *instanceTemplateShowCmd) CmdAliases() []string { return GShowAlias }

func (c *instanceTemplateShowCmd) CmdShort() string {
	return "Show a Compute instance template details"
}

func (c *instanceTemplateShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance template details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceTemplateShowOutput{}), ", "))
}

func (c *instanceTemplateShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		GContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	template, err := globalstate.EgoscaleClient.FindTemplate(ctx, c.Zone, c.Template, c.Visibility)
	if err != nil {
		return fmt.Errorf(
			"no template %q found with visibility %s in zone %s",
			c.Template,
			c.Visibility,
			c.Zone,
		)
	}

	return c.OutputFunc(&instanceTemplateShowOutput{
		ID:              *template.ID,
		Zone:            c.Zone,
		Family:          utils.DefaultString(template.Family, ""),
		Name:            *template.Name,
		Description:     utils.DefaultString(template.Description, ""),
		CreationDate:    template.CreatedAt.String(),
		Visibility:      *template.Visibility,
		Size:            *template.Size,
		Version:         utils.DefaultString(template.Version, ""),
		Build:           utils.DefaultString(template.Build, ""),
		Maintainer:      utils.DefaultString(template.Maintainer, ""),
		Checksum:        *template.Checksum,
		DefaultUser:     utils.DefaultString(template.DefaultUser, ""),
		SSHKeyEnabled:   *template.SSHKeyEnabled,
		PasswordEnabled: *template.PasswordEnabled,
		BootMode:        *template.BootMode,
	}, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceTemplateCmd, &instanceTemplateShowCmd{
		CliCommandSettings: DefaultCLICmdSettings(),

		Visibility: defaultTemplateVisibility,
	}))
}
