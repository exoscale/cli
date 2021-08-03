package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeSSHKeyShowOutput struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

func (o *computeSSHKeyShowOutput) toJSON()  { outputJSON(o) }
func (o *computeSSHKeyShowOutput) toText()  { outputText(o) }
func (o *computeSSHKeyShowOutput) toTable() { outputTable(o) }

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
		strings.Join(outputterTemplateAnnotations(&computeSSHKeyShowOutput{}), ", "))
}

func (c *computeSSHKeyShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	sshKey, err := cs.Client.GetSSHKey(ctx, gCurrentAccount.DefaultZone, c.Key)
	if err != nil {
		return err
	}

	return output(&computeSSHKeyShowOutput{
		Name:        *sshKey.Name,
		Fingerprint: *sshKey.Fingerprint,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
