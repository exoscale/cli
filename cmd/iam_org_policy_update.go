package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamOrgPolicyUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	Policy string `cli-arg:"#"`

	_ bool `cli-cmd:"update"`
}

func (c *iamOrgPolicyUpdateCmd) cmdAliases() []string {
	return []string{"replace"}
}

func (c *iamOrgPolicyUpdateCmd) cmdShort() string {
	return "Update Org policy"
}

func (c *iamOrgPolicyUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command replaces the complete IAM Organization Policy with the new one provided in JSON format.
To read the Policy from STDIN provide '-' as an argument.

Pro Tip: you can get the policy in JSON format with the command:

	exo iam org-policy show --output-format json

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamPolicyOutput{}), ", "))
}

func (c *iamOrgPolicyUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamOrgPolicyUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	if c.Policy == "-" {
		inputReader := cmd.InOrStdin()
		b, err := io.ReadAll(inputReader)
		if err != nil {
			return fmt.Errorf("failed to read policy from stdin: %w", err)
		}

		c.Policy = string(b)
	}

	policy, err := iamPolicyFromJSON([]byte(c.Policy))
	if err != nil {
		return fmt.Errorf("failed to parse IAM policy: %w", err)
	}

	op, err := client.UpdateIAMOrganizationPolicy(ctx, *policy)
	if err != nil {
		return err
	}
	decorateAsyncOperation("Updating IAM org policy...", func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&iamOrgPolicyShowCmd{
			cliCommandSettings: c.cliCommandSettings,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamOrgPolicyCmd, &iamOrgPolicyUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
