package iam

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamRoleCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name             string            `cli-arg:"#" cli-usage:"NAME"`
	Description      string            `cli-flag:"description" cli-usage:"Role description"`
	Permissions      []string          `cli-flag:"permissions" cli-usage:"Role permissions"`
	Editable         bool              `cli-flag:"editable" cli-usage:"Set --editable=false do prevent editing Policy after creation"`
	Labels           map[string]string `cli-flag:"label" cli-usage:"Role labels (format: key=value)"`
	Policy           string            `cli-flag:"policy" cli-usage:"Role policy (use '-' to read from STDIN)"`
	AssumeRolePolicy string            `cli-flag:"assume-role-policy" cli-usage:"Assume Role policy (use '-' to read from STDIN)"`
	MaxSessionTtl    int64             `cli-flag:"max-session-ttl" cli-usage:"Maximum TTL requester is allowed to ask for when assuming a role (0 implies default)"`
}

func (c *iamRoleCreateCmd) CmdAliases() []string { return nil }

func (c *iamRoleCreateCmd) CmdShort() string {
	return "Create IAM Role"
}

func (c *iamRoleCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a new IAM Role.
To read a policy from STDIN, append '-' to the '--policy' or '--assume-role-policy' flag.

Pro Tip: you can reuse an existing policy by providing the output of the show command as input:

	exo iam role show --policy --output-format json <role-name> | exo iam role create <new-role-name> --policy -

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamRoleShowOutput{}), ", "))
}

func (c *iamRoleCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleCreateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	if c.Name == "" {
		return errors.New("NAME not provided")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	var policy *v3.IAMPolicy
	var assumeRolePolicy *v3.IAMPolicy

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

	if c.AssumeRolePolicy != "" {
		// If Assume Role Policy value is `-` read from STDIN
		if c.AssumeRolePolicy == "-" {
			inputReader := cmd.InOrStdin()
			b, err := io.ReadAll(inputReader)
			if err != nil {
				return fmt.Errorf("failed to read assume role policy from stdin: %w", err)
			}

			c.AssumeRolePolicy = string(b)
		}

		var err error

		assumeRolePolicy, err = iamPolicyFromJSON([]byte(c.AssumeRolePolicy))
		if err != nil {
			return fmt.Errorf("failed to parse IAM policy: %w", err)
		}
	}

	role := v3.CreateIAMRoleRequest{
		Name:             c.Name,
		Editable:         &c.Editable,
		Labels:           c.Labels,
		Permissions:      c.Permissions,
		Policy:           policy,
		AssumeRolePolicy: assumeRolePolicy,
	}

	if c.Description != "" {
		role.Description = c.Description
	}

	if c.MaxSessionTtl != 0 {
		role.MaxSessionTtl = c.MaxSessionTtl
	}

	op, err := client.CreateIAMRole(ctx, role)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation("Creating IAM role...", func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&iamRoleShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Role:               op.Reference.ID.String(),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamRoleCmd, &iamRoleCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
		Editable:           true,
	}))
}
