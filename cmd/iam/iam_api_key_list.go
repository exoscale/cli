package iam

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
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

	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	rolesMap := map[string]string{}

	// For better UX we will print both role name and ID
	listIAMRolesResp, err := client.ListIAMRoles(ctx)
	// If API returns error, can continue (print UUID only) as this is non-essential feature
	if err == nil {
		for _, role := range listIAMRolesResp.IAMRoles {
			if role.ID.String() != "" && role.Name != "" {
				rolesMap[role.ID.String()] = role.Name
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
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *iamAPIKeyListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *iamAPIKeyListCmd) CmdShort() string { return "List API Keys" }

func (c *iamAPIKeyListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists all API Keys.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamAPIKeyListOutput{}), ", "))
}

func (c *iamAPIKeyListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAPIKeyListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	listAPIKeysResp, err := client.ListAPIKeys(ctx)
	if err != nil {
		return err
	}

	out := make(iamAPIKeyListOutput, 0)

	for _, apikey := range listAPIKeysResp.APIKeys {
		out = append(out, iamAPIKeyListItemOutput{
			Name: apikey.Name,
			Key:  apikey.Key,
			Role: apikey.RoleID.String(),
		})
	}

	return c.OutputFunc(&out, err)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamAPIKeyCmd, &iamAPIKeyListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
