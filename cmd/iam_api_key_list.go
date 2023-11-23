package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamAPIKeyListItemOutput struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	Role string `json:"role-id"`
}

type iamAPIKeyListOutput []iamAPIKeyListItemOutput

func (o *iamAPIKeyListOutput) ToJSON() { output.JSON(o) }
func (o *iamAPIKeyListOutput) ToText() { output.Text(o) }
func (o *iamAPIKeyListOutput) ToTable() {
	t := table.NewTable(os.Stdout)

	t.SetHeader([]string{
		"NAME",
		"KEY",
		"ROLE",
	})
	defer t.Render()

	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	// For better UX we will print both role name and ID
	rolesMap := map[string]string{}
	iamRoles, err := globalstate.EgoscaleClient.ListIAMRoles(ctx, zone)
	// If API returns error, can continue (print name only) as this is non-essential feature
	if err == nil {
		for _, role := range iamRoles {
			if role.ID != nil && role.Name != nil {
				rolesMap[*role.ID] = *role.Name
			}
		}
	}

	for _, apikey := range *o {
		role := apikey.Role
		if name, ok := rolesMap[apikey.Role]; ok {
			role = fmt.Sprintf("%s (%s)", name, apikey.Role)
		}

		t.Append([]string{
			apikey.Name,
			apikey.Key,
			role,
		})
	}
}

type iamAPIKeyListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *iamAPIKeyListCmd) cmdAliases() []string { return gListAlias }

func (c *iamAPIKeyListCmd) cmdShort() string { return "List API Keys" }

func (c *iamAPIKeyListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists all API Keys.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamAPIKeyListOutput{}), ", "))
}

func (c *iamAPIKeyListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAPIKeyListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	apikeys, err := globalstate.EgoscaleClient.ListAPIKeys(ctx, zone)
	if err != nil {
		return err
	}

	out := make(iamAPIKeyListOutput, 0)

	for _, apikey := range apikeys {
		out = append(out, iamAPIKeyListItemOutput{
			Name: utils.DefaultString(apikey.Name, ""),
			Key:  utils.DefaultString(apikey.Key, ""),
			Role: utils.DefaultString(apikey.RoleID, ""),
		})
	}

	return c.outputFunc(&out, err)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAPIKeyCmd, &iamAPIKeyListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
