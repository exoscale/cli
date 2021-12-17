package cmd

import (
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

type dbaasIntegrationDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ID string `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasIntegrationDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *dbaasIntegrationDeleteCmd) cmdShort() string { return "Delete a Database Service integration" }

func (c *dbaasIntegrationDeleteCmd) cmdLong() string { return "" }

func (c *dbaasIntegrationDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasIntegrationDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Database Service integration %q?", c.ID)) {
			return nil
		}
	}

	var (
		res *oapi.DeleteDbaasIntegrationResponse
		err error
	)
	decorateAsyncOperation(fmt.Sprintf("Deleting Database Service integration %q...", c.ID), func() {
		res, err = cs.DeleteDbaasIntegrationWithResponse(ctx, c.ID)
	})
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasIntegrationCmd, &dbaasIntegrationDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
