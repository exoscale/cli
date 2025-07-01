package instance

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSnapshotListItemOutput struct {
	ID           string `json:"id"`
	CreationDate string `json:"creation_date"`
	Instance     string `json:"instance"`
	State        string `json:"state"`
	Zone         string `json:"zone"`
}

type instanceSnapshotListOutput []instanceSnapshotListItemOutput

func (o *instanceSnapshotListOutput) ToJSON()  { output.JSON(o) }
func (o *instanceSnapshotListOutput) ToText()  { output.Text(o) }
func (o *instanceSnapshotListOutput) ToTable() { output.Table(o) }

type instanceSnapshotListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *instanceSnapshotListCmd) CmdAliases() []string { return nil }

func (c *instanceSnapshotListCmd) CmdShort() string { return "List Compute instance snapshots" }

func (c *instanceSnapshotListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists existing Compute instance snapshots.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceSnapshotListOutput{}), ", "))
}

func (c *instanceSnapshotListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var zones []v3.ZoneName
	ctx := exocmd.GContext

	if c.Zone != "" {
		zones = []v3.ZoneName{v3.ZoneName(c.Zone)}
	} else {
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
		if err != nil {
			return err
		}
		zones, err = utils.AllZonesV3(ctx, client)
		if err != nil {
			return err
		}
	}

	out := make(instanceSnapshotListOutput, 0)
	res := make(chan instanceSnapshotListItemOutput)
	done := make(chan struct{})

	var instances sync.Map

	go func() {
		for dt := range res {
			out = append(out, dt)
		}
		done <- struct{}{}
	}()
	err := utils.ForEachZone(zones, func(zone v3.ZoneName) error {
		ctx := exocmd.GContext
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}

		snapshots, err := client.ListSnapshots(ctx)
		if err != nil {
			return err
		}

		for _, s := range snapshots.Snapshots {
			var instance *v3.Instance
			instanceI, cached := instances.Load(s.Instance.ID.String())
			if cached {
				instance = instanceI.(*v3.Instance)
			} else {
				instance, err = client.GetInstance(ctx, s.Instance.ID)
				if err != nil {
					return fmt.Errorf("unable to retrieve Compute instance %q: %w", s.Instance.ID.String(), err)
				}
				instances.Store(s.Instance.ID.String(), instance)
			}

			res <- instanceSnapshotListItemOutput{
				ID:           s.ID.String(),
				CreationDate: s.CreatedAT.String(),
				Instance:     instance.Name,
				State:        string(s.State),
				Zone:         string(zone),
			}
		}

		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	close(res)
	<-done

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
