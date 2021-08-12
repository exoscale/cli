package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbServiceBackupListItemOutput struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
	Size int64     `json:"size"`
}

type dbServiceBackupListOutput []dbServiceBackupListItemOutput

func (o *dbServiceBackupListOutput) toJSON()  { outputJSON(o) }
func (o *dbServiceBackupListOutput) toText()  { outputText(o) }
func (o *dbServiceBackupListOutput) toTable() { outputTable(o) }

type dbServiceMaintenanceShowOutput struct {
	DOW  string `json:"dow"`
	Time string `json:"time"`
}

type dbServiceUserShowOutput struct {
	Type     string
	UserName string
}

type dbServiceShowOutput struct {
	Name                  string                          `json:"name"`
	Type                  string                          `json:"type"`
	Plan                  string                          `json:"plan"`
	CreationDate          time.Time                       `json:"creation_date"`
	Nodes                 int64                           `json:"nodes"`
	NodeCPUs              int64                           `json:"node_cpus"`
	NodeMemory            int64                           `json:"node_memory"`
	UpdateDate            time.Time                       `json:"update_date"`
	DiskSize              int64                           `json:"disk_size"`
	State                 string                          `json:"state"`
	TerminationProtection bool                            `json:"termination_protection"`
	Maintenance           *dbServiceMaintenanceShowOutput `json:"maintenance"`
	Users                 []dbServiceUserShowOutput       `json:"users"`
	Features              map[string]interface{}          `json:"features"`
	Metadata              map[string]interface{}          `json:"metadata"`
	Zone                  string                          `json:"zone"`
}

func (o *dbServiceShowOutput) toJSON() { outputJSON(o) }
func (o *dbServiceShowOutput) toText() { outputText(o) }
func (o *dbServiceShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Database Service"})
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"Type", o.Type})
	t.Append([]string{"Plan", o.Plan})
	t.Append([]string{"Creation Date", fmt.Sprint(o.CreationDate)})
	t.Append([]string{"Nodes", fmt.Sprint(o.Nodes)})
	t.Append([]string{"Node CPUs", fmt.Sprint(o.NodeCPUs)})
	t.Append([]string{"Node Memory", humanize.Bytes(uint64(o.NodeMemory))})
	t.Append([]string{"Update Date", fmt.Sprint(o.UpdateDate)})
	t.Append([]string{"Disk Size", humanize.Bytes(uint64(o.DiskSize))})
	t.Append([]string{"State", o.State})
	t.Append([]string{"Termination Protected", fmt.Sprint(o.TerminationProtection)})

	t.Append([]string{"Maintenance", func() string {
		if o.Maintenance != nil {
			return fmt.Sprintf("%s (%s)", o.Maintenance.DOW, o.Maintenance.Time)
		}
		return "n/a"
	}()})

	t.Append([]string{"Users", func() string {
		if len(o.Users) > 0 {
			return strings.Join(
				func() []string {
					users := make([]string, len(o.Users))
					for i := range o.Users {
						users[i] = fmt.Sprintf("%s (%s)", o.Users[i].UserName, o.Users[i].Type)
					}
					return users
				}(),
				"\n")
		}
		return "n/a"
	}()})

	t.Append([]string{"Features", func() string {
		sortedKeys := func() []string {
			keys := make([]string, 0)
			for k := range o.Features {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return keys
		}()

		buf := bytes.NewBuffer(nil)
		ft := table.NewEmbeddedTable(buf)
		ft.SetHeader([]string{" "})
		for _, k := range sortedKeys {
			ft.Append([]string{k, fmt.Sprint(o.Features[k])})
		}
		ft.Render()

		return buf.String()
	}()})

	t.Append([]string{"Metadata", func() string {
		sortedKeys := func() []string {
			keys := make([]string, 0)
			for k := range o.Metadata {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return keys
		}()

		buf := bytes.NewBuffer(nil)
		mt := table.NewEmbeddedTable(buf)
		mt.SetHeader([]string{" "})
		for _, k := range sortedKeys {
			mt.Append([]string{k, fmt.Sprint(o.Metadata[k])})
		}
		mt.Render()

		return buf.String()
	}()})
}

type dbServiceShowCmd struct {
	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#"`

	ShowBackups    bool   `cli-flag:"backups" cli-short:"b" cli-usage:"show Database Service backups"`
	ShowURI        bool   `cli-flag:"uri" cli-short:"u" cli-usage:"show Database Service connection URL"`
	ShowUserConfig bool   `cli-flag:"user-config" cli-short:"c" cli-usage:"show Database Service user config"`
	Zone           string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbServiceShowCmd) cmdAliases() []string { return gShowAlias }

func (c *dbServiceShowCmd) cmdShort() string { return "Show a Database Service details" }

func (c *dbServiceShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service details.

Supported output template annotations:

	* When showing a Database Service: %s
	* When showing a Database Service backups: %s`,
		strings.Join(outputterTemplateAnnotations(&dbServiceShowOutput{}), ", "),
		strings.Join(outputterTemplateAnnotations(&dbServiceBackupListItemOutput{}), ", "))
}

func (c *dbServiceShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbServiceShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if c.ShowBackups || c.ShowURI || c.ShowUserConfig {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

		dbService, err := cs.GetDatabaseService(ctx, c.Zone, c.Name)
		if err != nil {
			return err
		}

		switch {
		case c.ShowBackups:
			backups := make(dbServiceBackupListOutput, len(dbService.Backups))
			for i, b := range dbService.Backups {
				backups[i] = dbServiceBackupListItemOutput{
					Name: *b.Name,
					Date: *b.Date,
					Size: *b.Size,
				}
			}
			return output(&backups, nil)

		case c.ShowURI:
			fmt.Println(dbService.URI)

		case c.ShowUserConfig:
			userConfig, err := json.MarshalIndent(dbService.UserConfig, "", "  ")
			if err != nil {
				return fmt.Errorf("error unmarshaling user config: %s", err)
			}
			fmt.Println(string(userConfig))
		}

		return nil
	}

	return output(showDatabaseService(c.Zone, c.Name))
}

func showDatabaseService(zone, name string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	databaseService, err := cs.GetDatabaseService(ctx, zone, name)
	if err != nil {
		return nil, err
	}

	out := dbServiceShowOutput{
		Name:                  *databaseService.Name,
		Type:                  *databaseService.Type,
		Plan:                  *databaseService.Plan,
		CreationDate:          *databaseService.CreatedAt,
		Nodes:                 *databaseService.Nodes,
		NodeCPUs:              *databaseService.NodeCPUs,
		NodeMemory:            *databaseService.NodeMemory,
		UpdateDate:            *databaseService.UpdatedAt,
		DiskSize:              *databaseService.DiskSize,
		State:                 *databaseService.State,
		TerminationProtection: *databaseService.TerminationProtection,
		Maintenance: func() (v *dbServiceMaintenanceShowOutput) {
			if databaseService.Maintenance != nil {
				v = &dbServiceMaintenanceShowOutput{
					DOW:  databaseService.Maintenance.DOW,
					Time: databaseService.Maintenance.Time,
				}
			}
			return
		}(),
		Features: databaseService.Features,
		Metadata: databaseService.Metadata,
		Users: func() []dbServiceUserShowOutput {
			list := make([]dbServiceUserShowOutput, len(databaseService.Users))
			for i, u := range databaseService.Users {
				list[i] = dbServiceUserShowOutput{
					UserName: *u.UserName,
					Type:     *u.Type,
				}
			}
			return list
		}(),
		Zone: zone,
	}

	return &out, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbCmd, &dbServiceShowCmd{}))
}
