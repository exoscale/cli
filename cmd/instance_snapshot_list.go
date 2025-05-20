package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	CliCommandSettings `cli-cmd:"-"`

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
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = utils.AllZones
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
	err := utils.ForEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

		list, err := globalstate.EgoscaleClient.ListSnapshots(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Compute instance snapshots in zone %s: %w", zone, err)
		}

		for _, s := range list {
			var instance *egoscale.Instance
			instanceI, cached := instances.Load(*s.InstanceID)
			if cached {
				instance = instanceI.(*egoscale.Instance)
			} else {
				instance, err = globalstate.EgoscaleClient.GetInstance(ctx, zone, *s.InstanceID)
				if err != nil {
					return fmt.Errorf("unable to retrieve Compute instance %q: %w", *s.InstanceID, err)
				}
				instances.Store(*s.InstanceID, instance)
			}

			res <- instanceSnapshotListItemOutput{
				ID:           *s.ID,
				CreationDate: s.CreatedAt.String(),
				Instance:     *instance.Name,
				State:        *s.State,
				Zone:         zone,
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
	cobra.CheckErr(RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotListCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
