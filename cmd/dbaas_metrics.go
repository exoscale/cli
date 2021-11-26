package cmd

import (
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

type dbaasServiceMetricsCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"metrics"`

	Name string `cli-arg:"#"`

	Period string `cli-usage:"metrics time period to retrieve"`
	Zone   string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceMetricsCmd) cmdAliases() []string { return gShowAlias }

func (c *dbaasServiceMetricsCmd) cmdShort() string {
	return "Query a Database Service metrics over time"
}

func (c *dbaasServiceMetricsCmd) cmdLong() string {
	return `This command outputs a Database Service raw metrics for the specified time
period in JSON format.`
}

func (c *dbaasServiceMetricsCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceMetricsCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	res, err := cs.GetDbaasServiceMetricsWithResponse(
		ctx,
		c.Name,
		oapi.GetDbaasServiceMetricsJSONRequestBody{Period: (*oapi.GetDbaasServiceMetricsJSONBodyPeriod)(&c.Period)},
	)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	fmt.Println(string(res.Body))

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasServiceMetricsCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Period: "hour",
	}))
}
