package iam

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamAPIKeyShowOutput struct {
	Name   string  `json:"name"`
	Key    string  `json:"key"`
	Secret string  `json:"secret"`
	Role   v3.UUID `json:"role-id"`
}

func (o *iamAPIKeyShowOutput) ToJSON()  { output.JSON(o) }
func (o *iamAPIKeyShowOutput) ToText()  { output.Text(o) }
func (o *iamAPIKeyShowOutput) ToTable() { output.Table(o) }

type iamAPIKeyCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
	Role string `cli-arg:"#" cli-usage:"ROLE-NAME|ROLE-ID"`
}

func (c *iamAPIKeyCreateCmd) CmdAliases() []string { return nil }

func (c *iamAPIKeyCreateCmd) CmdShort() string {
	return "Create API Key"
}

func (c *iamAPIKeyCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a new API Key.
Because Secret is only printed during API Key creation, --quiet (-Q) flag is not implemented for this command.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamAPIKeyShowOutput{}), ", "))
}

func (c *iamAPIKeyCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAPIKeyCreateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	listIAMRolesResp, err := client.ListIAMRoles(ctx)
	if err != nil {
		return err
	}

	iamRole, err := listIAMRolesResp.FindIAMRole(c.Role)
	if err != nil {
		return err
	}

	createAPIKeyReq := v3.CreateAPIKeyRequest{
		Name:   c.Name,
		RoleID: iamRole.ID,
	}

	iamAPIKey, err := client.CreateAPIKey(ctx, createAPIKeyReq)
	if err != nil {
		return err
	}

	out := iamAPIKeyShowOutput{
		Name:   iamAPIKey.Name,
		Key:    iamAPIKey.Key,
		Secret: iamAPIKey.Secret,
		Role:   iamAPIKey.RoleID,
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamAPIKeyCmd, &iamAPIKeyCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
