package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeSSHKeyListItemOutput struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
}

type computeSSHKeyListOutput []computeSSHKeyListItemOutput

func (o *computeSSHKeyListOutput) toJSON()  { outputJSON(o) }
func (o *computeSSHKeyListOutput) toText()  { outputText(o) }
func (o *computeSSHKeyListOutput) toTable() { outputTable(o) }

type computeSSHKeyListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *computeSSHKeyListCmd) cmdAliases() []string { return nil }

func (c *computeSSHKeyListCmd) cmdShort() string { return "List SSH keys" }

func (c *computeSSHKeyListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists SSH keys.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&computeSSHKeyListItemOutput{}), ", "))
}

func (c *computeSSHKeyListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	sshKeys, err := cs.ListSSHKeys(ctx, gCurrentAccount.DefaultZone)
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

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
