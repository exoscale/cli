package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamRoleCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name        string            `cli-arg:"#" cli-usage:"NAME"`
	Description string            `cli-flag:"description" cli-usage:"Role description"`
	Permissions []string          `cli-flag:"permissions" cli-usage:"Role permissions"`
	Editable    bool              `cli-flag:"editable" cli-usage:"Set --editable=false do prevent editing Policy after creation"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Role labels (format: key=value)"`
	Policy      string            `cli-flag:"policy" cli-usage:"Role policy (use '-' to read from STDIN)"`
}

func (c *iamRoleCreateCmd) cmdAliases() []string { return nil }

func (c *iamRoleCreateCmd) cmdShort() string {
	return "Create IAM Role"
}

func (c *iamRoleCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a new IAM Role.
To read the Policy from STDIN, append '-' to the '--policy' flag.

Pro Tip: you can reuse an existing role policy by providing the output of the show command as input:

	exo iam role show --policy --output-format json <role-name> | exo iam role create <new-role-name> --policy -

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamRoleShowOutput{}), ", "))
}

func (c *iamRoleCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleCreateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if c.Name == "" {
		return errors.New("NAME not provided")
	}

	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	var policy *v3.IAMPolicy

	// Policy is optional, if not set API will default to `allow all`
	if c.Policy != "" {
		// If Policy value is `-` read from STDIN
		if c.Policy == "-" {
			inputReader := cmd.InOrStdin()
			b, err := io.ReadAll(inputReader)
			if err != nil {
				return fmt.Errorf("failed to read policy from stdin: %w", err)
			}

			c.Policy = string(b)
		}

		var err error

		policy, err = iamPolicyFromJSON([]byte(c.Policy))
		if err != nil {
			return fmt.Errorf("failed to parse IAM policy: %w", err)
		}
	}

	role := v3.CreateIAMRoleRequest{
		Name:        c.Name,
		Editable:    &c.Editable,
		Labels:      c.Labels,
		Permissions: c.Permissions,
		Policy:      policy,
	}

	if c.Description != "" {
		role.Description = c.Description
	}

	op, err := client.CreateIAMRole(ctx, role)
	if err != nil {
		return err
	}
	decorateAsyncOperation("Creating IAM role...", func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&iamRoleShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Role:               op.Reference.ID.String(),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamRoleCmd, &iamRoleCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
		Editable:           true,
	}))
}
