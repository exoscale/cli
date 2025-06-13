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
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"metrics"`

	Name string `cli-arg:"#"`

	Period string `cli-usage:"metrics time period to retrieve"`
	Zone   string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceMetricsCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *dbaasServiceMetricsCmd) CmdShort() string {
	return "Query a Database Service metrics over time"
}

func (c *dbaasServiceMetricsCmd) CmdLong() string {
	return `This command outputs a Database Service raw metrics for the specified time
period in JSON format.`
}

func (c *dbaasServiceMetricsCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

<<<<<<< Updated upstream:cmd/dbaas_metrics.go
func (c *dbaasServiceMetricsCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
	res, err := client.GetDBAASServiceMetrics(
=======
func (c *dbaasServiceMetricsCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	res, err := globalstate.EgoscaleClient.GetDbaasServiceMetricsWithResponse(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_metrics.go
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
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceMetricsCmd{
		cliCommandSettings: exocmd.DefaultCLICmdSettings(),

		Period: "hour",
	}))
}
