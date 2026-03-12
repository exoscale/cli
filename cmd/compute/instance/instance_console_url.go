package instance

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type InstanceConsoleURLOutput struct {
	ConsoleURL string `json:"console-url"`
}

func (o *InstanceConsoleURLOutput) Type() string { return "Compute instance" }
func (o *InstanceConsoleURLOutput) ToJSON()      { output.JSON(o) }
func (o *InstanceConsoleURLOutput) ToText()      { output.Text(o) }
func (o *InstanceConsoleURLOutput) ToTable()     { output.Table(o) }

type instanceConsoleURLCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"console-url"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceConsoleURLCmd) CmdAliases() []string { return nil }

func (c *instanceConsoleURLCmd) CmdShort() string { return "Get instance console URL" }

func (c *instanceConsoleURLCmd) CmdLong() string {
	return fmt.Sprintf(`This command generates a Compute instance console URL.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceConsoleURLOutput{}), ", "))
}

func (c *instanceConsoleURLCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceConsoleURLCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}

	foundInstance, err := findInstance(resp, c.Instance, string(c.Zone))
	if err != nil {
		return err
	}

	consoleProxyURL, err := client.GetConsoleProxyURL(ctx, foundInstance.ID)

	if err != nil {
		return err
	}

	prefix, _, found := strings.Cut(consoleProxyURL.Host, "console-")
	if !found {
		prefix = ""
	}

	u := &url.URL{
		Scheme: "https",
		Host:   prefix + "portal.exoscale.com",
		Path:   "vnc",
	}

	q := u.Query()
	q.Set("host", consoleProxyURL.Host)
	q.Set("path", consoleProxyURL.Path)
	u.RawQuery = q.Encode()

	out := InstanceConsoleURLOutput{
		ConsoleURL: u.String(),
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceConsoleURLCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
