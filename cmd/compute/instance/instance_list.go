package instance

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

const (
	EmptyIPAddressVisualization = "-"
)

type instanceListItemOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Zone      string `json:"zone"`
	Type      string `json:"type"`
	IPAddress string `json:"ip_address"`
	State     string `json:"state"`
}

type instanceListOutput []instanceListItemOutput

func (o *instanceListOutput) ToJSON()  { output.JSON(o) }
func (o *instanceListOutput) ToText()  { output.Text(o) }
func (o *instanceListOutput) ToTable() { output.Table(o) }

type instanceListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *instanceListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *instanceListCmd) CmdShort() string { return "List Compute instances" }

func (c *instanceListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Compute instances.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceListItemOutput{}), ", "))
}

func (c *instanceListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = utils.AllZones
	}

	out := make(instanceListOutput, 0)
	res := make(chan instanceListItemOutput)
	done := make(chan struct{})

	var instanceTypes sync.Map

	go func() {
		for instance := range res {
			out = append(out, instance)
		}
		done <- struct{}{}
	}()
	err := utils.ForEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

		list, err := globalstate.EgoscaleClient.ListInstances(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Compute instances in zone %s: %w", zone, err)
		}

		for _, i := range list {
			var instanceType *egoscale.InstanceType
			instanceTypeI, cached := instanceTypes.Load(*i.InstanceTypeID)
			if cached {
				instanceType = instanceTypeI.(*egoscale.InstanceType)
			} else {
				instanceType, err = globalstate.EgoscaleClient.GetInstanceType(ctx, zone, *i.InstanceTypeID)
				if err != nil {
					return fmt.Errorf(
						"unable to retrieve Compute instance type %q: %w",
						*i.InstanceTypeID,
						err)
				}
				instanceTypes.Store(*i.InstanceTypeID, instanceType)
			}

			res <- instanceListItemOutput{
				ID:        *i.ID,
				Name:      *i.Name,
				Zone:      zone,
				Type:      fmt.Sprintf("%s.%s", *instanceType.Family, *instanceType.Size),
				IPAddress: utils.DefaultIP(i.PublicIPAddress, EmptyIPAddressVisualization),
				State:     *i.State,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
