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
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"metrics"`

	Name string `cli-arg:"#"`

	Period string `cli-usage:"metrics time period to retrieve"`
	Zone   string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceMetricsCmd) CmdAliases() []string { return GShowAlias }

func (c *dbaasServiceMetricsCmd) CmdShort() string {
	return "Query a Database Service metrics over time"
}

func (c *dbaasServiceMetricsCmd) CmdLong() string {
	return `This command outputs a Database Service raw metrics for the specified time
period in JSON format.`
}

func (c *dbaasServiceMetricsCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceMetricsCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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
	cobra.CheckErr(RegisterCLICommand(dbaasCmd, &dbaasServiceMetricsCmd{
		CliCommandSettings: DefaultCLICmdSettings(),

		Period: "hour",
	}))
}
