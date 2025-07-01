package instance_pool

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instancePoolListItemOutput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Size  int64  `json:"size"`
	State string `json:"state"`
}

type instancePoolListOutput []instancePoolListItemOutput

func (o *instancePoolListOutput) ToJSON()  { output.JSON(o) }
func (o *instancePoolListOutput) ToText()  { output.Text(o) }
func (o *instancePoolListOutput) ToTable() { output.Table(o) }

type instancePoolListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *instancePoolListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *instancePoolListCmd) CmdShort() string { return "List Instance Pools" }

func (c *instancePoolListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Instance Pools.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instancePoolListItemOutput{}), ", "))
}

func (c *instancePoolListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolListCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	out := make(instancePoolListOutput, 0)
	res := make(chan instancePoolListItemOutput)
	done := make(chan struct{})

	go func() {
		for instancePool := range res {
			out = append(out, instancePool)
		}
		done <- struct{}{}
	}()
	err := utils.ForEachZone(zones, func(zone v3.ZoneName) error {
		ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, string(zone)))

		list, err := globalstate.EgoscaleClient.ListInstancePools(ctx, string(zone))
		if err != nil {
			return fmt.Errorf("unable to list Instance Pools in zone %s: %w", zone, err)
		}

		for _, i := range list {
			res <- instancePoolListItemOutput{
				ID:    *i.ID,
				Name:  *i.Name,
				Zone:  string(zone),
				Size:  *i.Size,
				State: *i.State,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePoolCmd, &instancePoolListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
