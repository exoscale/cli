package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbServiceLogsItemOutput struct {
	Time    time.Time `json:"time"`
	Node    string    `json:"node"`
	Unit    string    `json:"unit"`
	Message string    `json:"message"`
}

type dbServiceLogsOutput struct {
	FirstLogOffset string                    `json:"first_log_offset"`
	Offset         string                    `json:"offset"`
	Logs           []dbServiceLogsItemOutput `json:"logs"`
}

func (o *dbServiceLogsOutput) ToJSON() { output.JSON(o) }
func (o *dbServiceLogsOutput) ToText() { output.Text(o) }
func (o *dbServiceLogsOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"Time", "Node", "Unit", "Message"})
	for _, notification := range o.Logs {
		t.Append([]string{
			notification.Time.String(),
			notification.Node,
			notification.Unit,
			notification.Message,
		})
	}
}

type dbaasServiceLogsCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"logs"`

	Name string `cli-arg:"#"`

	Limit  int64  `cli-short:"l" cli-usage:"number of log messages to retrieve"`
	Offset string `cli-short:"o" cli-usage:"opaque offset identifier (can be found in the JSON output of the command)"`
	Sort   string `cli-usage:"log messages sorting order (asc|desc)"`
	Zone   string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceLogsCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *dbaasServiceLogsCmd) CmdShort() string {
	return "Query a Database Service logs"
}

func (c *dbaasServiceLogsCmd) CmdLong() string {
	return fmt.Sprintf(`This command outputs a Database Service logs.

Supported output template annotations: %s
  - .Logs: %s

Example usage with custom output containing only the actual log messages:

    exo dbaas logs MY-SERVICE --output-template \
        '{{range $l := .Logs}}{{println $l.Message}}{{end}}'
`,
		strings.Join(output.TemplateAnnotations(&dbServiceLogsOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceLogsItemOutput{}), ", "))
}

func (c *dbaasServiceLogsCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

<<<<<<< Updated upstream:cmd/dbaas_logs.go
func (c *dbaasServiceLogsCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
=======
func (c *dbaasServiceLogsCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_logs.go

	svcLogRequest := v3.GetDBAASServiceLogsRequest{}
	if cmd.Flags().Changed("limit") {
		svcLogRequest.Limit = c.Limit
	}
	if cmd.Flags().Changed("offset") {
		svcLogRequest.Offset = c.Offset
	}
	if cmd.Flags().Changed("sort") {
		svcLogRequest.SortOrder = v3.EnumSortOrder(c.Sort)
	}

	res, err := client.GetDBAASServiceLogs(
		ctx,
		c.Name,
		svcLogRequest,
	)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	out := dbServiceLogsOutput{
		FirstLogOffset: utils.DefaultString(&res.FirstLogOffset, "-"),
		Offset:         utils.DefaultString(&res.Offset, "-"),
		Logs:           make([]dbServiceLogsItemOutput, len(res.Logs)),
	}

	for i, log := range res.Logs {
		ts, err := time.Parse("2006-01-02T15:04:05.000000", log.Time)
		if err != nil {
			return fmt.Errorf("unable to parse log timestamp: %w", err)
		}
		out.Logs[i].Time = ts

		out.Logs[i].Node = utils.DefaultString(&log.Node, "-")
		out.Logs[i].Unit = utils.DefaultString(&log.Unit, "-")
		out.Logs[i].Message = utils.DefaultString(&log.Message, "-")
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceLogsCmd{
		cliCommandSettings: exocmd.DefaultCLICmdSettings(),

		Limit: 10,
		Sort:  "desc",
	}))
}
