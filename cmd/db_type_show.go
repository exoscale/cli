package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type dbTypePlanShowOutput struct {
	Name       string
	Nodes      int64
	NodeCPUs   int64
	NodeMemory int64
	DiskSpace  int64
}

type dbTypeShowOutput struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	LatestVersion  string                 `json:"latest_version"`
	DefaultVersion string                 `json:"default_version"`
	Plans          []dbTypePlanShowOutput `json:"plans"`
}

func (o *dbTypeShowOutput) toJSON() { outputJSON(o) }
func (o *dbTypeShowOutput) toText() { outputText(o) }
func (o *dbTypeShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Latest Version", o.LatestVersion})
	t.Append([]string{"Default Version", o.DefaultVersion})

	t.Append([]string{"Plans", func() string {
		entries := make([]string, len(o.Plans))
		for _, p := range o.Plans {
			buf := bytes.NewBuffer(nil)
			pt := table.NewEmbeddedTable(buf)
			pt.SetHeader([]string{" "})
			pt.Append([]string{"Name", p.Name})
			pt.Append([]string{"Nodes", fmt.Sprint(p.Nodes)})
			pt.Append([]string{"Node CPUs", fmt.Sprint(p.NodeCPUs)})
			pt.Append([]string{"Node Memory", humanize.Bytes(uint64(p.NodeMemory))})
			pt.Append([]string{"Disk Space", humanize.Bytes(uint64(p.DiskSpace))})
			pt.SetAlignment(tablewriter.ALIGN_RIGHT)
			pt.Render()
			entries = append(entries, buf.String())
		}
		return strings.Join(entries, "")
	}()})
}

type dbTypeShowCmd struct {
	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#"`
}

func (c *dbTypeShowCmd) cmdAliases() []string { return gShowAlias }

func (c *dbTypeShowCmd) cmdShort() string { return "Show a Database Service type details" }

func (c *dbTypeShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service type details.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&dbTypeShowOutput{}), ", "))
}

func (c *dbTypeShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbTypeShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	dt, err := cs.GetDatabaseServiceType(ctx, gCurrentAccount.DefaultZone, c.Name)
	if err != nil {
		return err
	}

	return output(&dbTypeShowOutput{
		Name:           *dt.Name,
		Description:    *dt.Description,
		LatestVersion:  *dt.LatestVersion,
		DefaultVersion: *dt.DefaultVersion,
		Plans: func() []dbTypePlanShowOutput {
			plans := make([]dbTypePlanShowOutput, len(dt.Plans))
			for i := range dt.Plans {
				plans[i] = dbTypePlanShowOutput{
					Name:       *dt.Plans[i].Name,
					Nodes:      *dt.Plans[i].Nodes,
					NodeCPUs:   *dt.Plans[i].NodeCPUs,
					NodeMemory: *dt.Plans[i].NodeMemory,
					DiskSpace:  *dt.Plans[i].DiskSpace,
				}
			}
			return plans
		}(),
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbTypeCmd, &dbTypeShowCmd{}))
}
