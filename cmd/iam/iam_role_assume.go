package iam

import (
	"errors"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamRoleAssumeOutput struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	OrgID  string `json:"org-id"`
	RoleID string `json:"role-id"`
	Secret string `json:"secret"`
}

func (o *iamRoleAssumeOutput) ToJSON()  { output.JSON(o) }
func (o *iamRoleAssumeOutput) ToText()  { output.Text(o) }
func (o *iamRoleAssumeOutput) ToTable() { output.Table(o) }

type iamRoleAssumeCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	Role string `cli-arg:"#" cli-usage:"ID|NAME"`

	Ttl int64 `cli-flag:"ttl" cli-usage:"Time To Live for the requested key in seconds (default: 300)"`

	_ bool `cli-cmd:"assume"`
}

func (c *iamRoleAssumeCmd) CmdAliases() []string { return nil }

func (c *iamRoleAssumeCmd) CmdShort() string {
	return "Request generation of key/secret allowing calls as of target role"
}

func (c *iamRoleAssumeCmd) CmdLong() string {
	return "Request generation of key/secret allowing calls as of target role"
}

func (c *iamRoleAssumeCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleAssumeCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	if c.Role == "" {
		return errors.New("role not provided")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	assumeRoleRq := v3.AssumeIAMRoleRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Ttl)) {
		assumeRoleRq.Ttl = c.Ttl
	}

	apiKey, err := client.AssumeIAMRole(ctx, v3.UUID(c.Role), assumeRoleRq)
	if err != nil {
		return err
	}

	out := iamRoleAssumeOutput{
		Key:    apiKey.Key,
		Name:   apiKey.Name,
		OrgID:  apiKey.OrgID,
		RoleID: apiKey.RoleID,
		Secret: apiKey.Secret,
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamRoleCmd, &iamRoleAssumeCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
