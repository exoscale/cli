package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/table"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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

func (o *instanceTemplateShowOutput) toJSON() { outputJSON(o) }
func (o *instanceTemplateShowOutput) toText() { outputText(o) }
func (o *instanceTemplateShowOutput) toTable() {
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Template string `cli-arg:"#" cli-usage:"[FAMILY.]SIZE"`

	Visibility string `cli-short:"v" cli-usage:"template visibility (public|private)"`
	Zone       string `cli-short:"z" cli-usage:"zone to filter results to (default: current account's default zone)"`
}

func (c *instanceTemplateShowCmd) cmdAliases() []string { return gShowAlias }

func (c *instanceTemplateShowCmd) cmdShort() string {
	return "Show a Compute instance template details"
}

func (c *instanceTemplateShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance template details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceTemplateShowOutput{}), ", "))
}

func (c *instanceTemplateShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	// Opportunistic shortcut in case the template is referenced by ID.
	template, _ := cs.Client.GetTemplate(ctx, c.Zone, c.Template)

	if template == nil {
		var templateID string

		templates, err := cs.ListTemplates(ctx, c.Zone, egoscale.ListTemplatesWithVisibility(c.Visibility))
		if err != nil {
			return fmt.Errorf("error retrieving templates: %w", err)
		}
		for _, template := range templates {
			if *template.ID == c.Template || *template.Name == c.Template {
				templateID = *template.ID
				break
			}
		}
		if templateID == "" {
			return fmt.Errorf(
				"no template %q found with visibility %s in zone %s",
				c.Template,
				c.Visibility,
				c.Zone,
			)
		}

		template, err = cs.Client.GetTemplate(ctx, c.Zone, templateID)
		if err != nil {
			return fmt.Errorf("error retrieving template: %w", err)
		}
	}

	return c.outputFunc(&instanceTemplateShowOutput{
		ID:              *template.ID,
		Zone:            c.Zone,
		Family:          defaultString(template.Family, ""),
		Name:            *template.Name,
		Description:     defaultString(template.Description, ""),
		CreationDate:    template.CreatedAt.String(),
		Visibility:      *template.Visibility,
		Size:            *template.Size,
		Version:         defaultString(template.Version, ""),
		Build:           defaultString(template.Build, ""),
		Maintainer:      defaultString(template.Maintainer, ""),
		Checksum:        *template.Checksum,
		DefaultUser:     defaultString(template.DefaultUser, ""),
		SSHKeyEnabled:   *template.SSHKeyEnabled,
		PasswordEnabled: *template.PasswordEnabled,
		BootMode:        *template.BootMode,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceTemplateCmd, &instanceTemplateShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Visibility: defaultTemplateVisibility,
	}))
}
