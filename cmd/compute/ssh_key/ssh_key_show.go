package ssh_key

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
)

type computeSSHKeyShowOutput struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

func (o *computeSSHKeyShowOutput) Type() string { return "SSH key" }
func (o *computeSSHKeyShowOutput) ToJSON()      { output.JSON(o) }
func (o *computeSSHKeyShowOutput) ToText()      { output.Text(o) }
func (o *computeSSHKeyShowOutput) ToTable()     { output.Table(o) }

type computeSSHKeyShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Key string `cli-arg:"#"`
}

func (c *computeSSHKeyShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *computeSSHKeyShowCmd) CmdShort() string {
	return "Show an SSH key details"
}

func (c *computeSSHKeyShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows an SSH key details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&computeSSHKeyShowOutput{}), ", "))
}

func (c *computeSSHKeyShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	sshKey, err := globalstate.EgoscaleV3Client.GetSSHKey(ctx, c.Key)
	if err != nil {
		return err
	}

	return c.OutputFunc(&computeSSHKeyShowOutput{
		Name:        sshKey.Name,
		Fingerprint: sshKey.Fingerprint,
	}, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(computeSSHKeyCmd, &computeSSHKeyShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
