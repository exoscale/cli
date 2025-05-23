package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
)

type computeSSHKeyListItemOutput struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

type computeSSHKeyListOutput []computeSSHKeyListItemOutput

func (o *computeSSHKeyListOutput) ToJSON()  { output.JSON(o) }
func (o *computeSSHKeyListOutput) ToText()  { output.Text(o) }
func (o *computeSSHKeyListOutput) ToTable() { output.Table(o) }

type computeSSHKeyListCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *computeSSHKeyListCmd) CmdAliases() []string { return nil }

func (c *computeSSHKeyListCmd) CmdShort() string { return "List SSH keys" }

func (c *computeSSHKeyListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists SSH keys.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&computeSSHKeyListItemOutput{}), ", "))
}

func (c *computeSSHKeyListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyListCmd) CmdRun(_ *cobra.Command, _ []string) error {

	ctx := GContext
	client := globalstate.EgoscaleV3Client

	sshKeysResponse, err := client.ListSSHKeys(ctx)
	if err != nil {
		return err
	}

	out := make(computeSSHKeyListOutput, 0)

	for _, k := range sshKeysResponse.SSHKeys {
		out = append(out, computeSSHKeyListItemOutput{
			Name:        k.Name,
			Fingerprint: k.Fingerprint,
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(computeSSHKeyCmd, &computeSSHKeyListCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
