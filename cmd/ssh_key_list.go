package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *computeSSHKeyListCmd) cmdAliases() []string { return nil }

func (c *computeSSHKeyListCmd) cmdShort() string { return "List SSH keys" }

func (c *computeSSHKeyListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists SSH keys.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&computeSSHKeyListItemOutput{}), ", "))
}

func (c *computeSSHKeyListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	sshKeys, err := globalstate.EgoscaleClient.ListSSHKeys(ctx, account.CurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := make(computeSSHKeyListOutput, 0)

	for _, k := range sshKeys {
		out = append(out, computeSSHKeyListItemOutput{
			Name:        *k.Name,
			Fingerprint: *k.Fingerprint,
		})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
