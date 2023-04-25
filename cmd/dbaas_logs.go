package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"logs"`

	Name string `cli-arg:"#"`

	Limit  int64  `cli-short:"l" cli-usage:"number of log messages to retrieve"`
	Offset string `cli-short:"o" cli-usage:"opaque offset identifier (can be found in the JSON output of the command)"`
	Sort   string `cli-usage:"log messages sorting order (asc|desc)"`
	Zone   string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceLogsCmd) cmdAliases() []string { return gShowAlias }

func (c *dbaasServiceLogsCmd) cmdShort() string {
	return "Query a Database Service logs"
}

func (c *dbaasServiceLogsCmd) cmdLong() string {
	return fmt.Sprintf(`This command outputs a Database Service logs.

Supported output template annotations: %s
  - .Logs: %s

Example usage with custom output containing only the actual log messages:

    exo dbaas logs MY-SERVICE --output-template \
        '{{range $l := .Logs}}{{println $l.Message}}{{end}}'
`,
		strings.Join(output.OutputterTemplateAnnotations(&dbServiceLogsOutput{}), ", "),
		strings.Join(output.OutputterTemplateAnnotations(&dbServiceLogsItemOutput{}), ", "))
}

func (c *dbaasServiceLogsCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceLogsCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	res, err := globalstate.GlobalEgoscaleClient.GetDbaasServiceLogsWithResponse(
		ctx,
		c.Name,
		oapi.GetDbaasServiceLogsJSONRequestBody{
			Limit:     &c.Limit,
			Offset:    utils.NonEmptyStringPtr(c.Offset),
			SortOrder: (*oapi.EnumSortOrder)(utils.NonEmptyStringPtr(c.Sort)),
		},
	)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	out := dbServiceLogsOutput{
		FirstLogOffset: utils.DefaultString(res.JSON200.FirstLogOffset, "-"),
		Offset:         utils.DefaultString(res.JSON200.Offset, "-"),
		Logs:           make([]dbServiceLogsItemOutput, len(*res.JSON200.Logs)),
	}

	for i, log := range *res.JSON200.Logs {
		ts, err := time.Parse("2006-01-02T15:04:05.000000", *log.Time)
		if err != nil {
			return fmt.Errorf("unable to parse log timestamp: %w", err)
		}
		out.Logs[i].Time = ts

		out.Logs[i].Node = utils.DefaultString(log.Node, "-")
		out.Logs[i].Unit = utils.DefaultString(log.Unit, "-")
		out.Logs[i].Message = utils.DefaultString(log.Message, "-")
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasServiceLogsCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		Limit: 10,
		Sort:  "desc",
	}))
}
