package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
	Role string `cli-arg:"#" cli-usage:"ROLE-NAME|ROLE-ID"`
}

func (c *iamAPIKeyCreateCmd) cmdAliases() []string { return nil }

func (c *iamAPIKeyCreateCmd) cmdShort() string {
	return "Create API Key"
}

func (c *iamAPIKeyCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a new API Key.
Because Secret is only printed during API Key creation, --quiet (-Q) flag is not implemented for this command.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamAPIKeyShowOutput{}), ", "))
}

func (c *iamAPIKeyCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAPIKeyCreateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := gContext
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

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAPIKeyCmd, &iamAPIKeyCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
