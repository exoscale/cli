package cmd

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamAPIKeyShowOutput struct {
	Name   string `json:"name"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
	Role   string `json:"role-id"`
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
	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone),
	)

	if _, err := uuid.Parse(c.Role); err != nil {
		roles, err := globalstate.EgoscaleClient.ListIAMRoles(ctx, zone)
		if err != nil {
			return err
		}

		for _, role := range roles {
			if role.Name != nil && *role.Name == c.Role {
				c.Role = *role.ID
				break
			}
		}
	}

	role, err := globalstate.EgoscaleClient.GetIAMRole(ctx, zone, c.Role)
	if err != nil {
		return err
	}

	apikey := &exoscale.APIKey{
		Name:   &c.Name,
		RoleID: role.ID,
	}

	k, secret, err := globalstate.EgoscaleClient.CreateAPIKey(ctx, zone, apikey)
	if err != nil {
		return err
	}

	out := iamAPIKeyShowOutput{
		Name:   utils.DefaultString(k.Name, ""),
		Key:    utils.DefaultString(k.Key, ""),
		Secret: secret,
		Role:   utils.DefaultString(k.RoleID, ""),
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAPIKeyCmd, &iamAPIKeyCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
