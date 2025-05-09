package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
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
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
	res, err := client.GetDBAASServiceMetrics(
		ctx,
		c.Name,
		v3.GetDBAASServiceMetricsRequest{
			Period: v3.GetDBAASServiceMetricsRequestPeriod(c.Period),
		},
	)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	out, err := json.Marshal(res.Metrics)
	if err != nil {
		return err
	}
	fmt.Println(string(out))

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasServiceMetricsCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Period: "hour",
	}))
}
