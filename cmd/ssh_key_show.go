package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Key string `cli-arg:"#"`
}

func (c *computeSSHKeyShowCmd) cmdAliases() []string { return gShowAlias }

func (c *computeSSHKeyShowCmd) cmdShort() string {
	return "Show an SSH key details"
}

func (c *computeSSHKeyShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows an SSH key details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&computeSSHKeyShowOutput{}), ", "))
}

func (c *computeSSHKeyShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	sshKey, err := globalstate.EgoscaleClient.Client.GetSSHKey(ctx, account.CurrentAccount.DefaultZone, c.Key)
	if err != nil {
		return err
	}

	return c.outputFunc(&computeSSHKeyShowOutput{
		Name:        *sshKey.Name,
		Fingerprint: *sshKey.Fingerprint,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
