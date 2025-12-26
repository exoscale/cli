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
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceListItemOutput struct {
	ID          v3.UUID     `json:"id"`
	Name        string      `json:"name"`
	Zone        v3.ZoneName `json:"zone"`
	Type        string      `json:"type"`
	IPAddress   string      `json:"ip_address"`
	IPv6Address string      `json:"ipv6_address"`
	State       string      `json:"state"`
}

type instanceListOutput []instanceListItemOutput

func (o *instanceListOutput) ToJSON() { output.JSON(o) }
func (o *instanceListOutput) ToText() { output.Text(o) }
func (o *instanceListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"ID",
		"NAME",
		"ZONE",
		"TYPE",
		"IP ADDRESS",
		"IPV6 ADDRESS",
		"STATE",
	})

	for _, instance := range *o {
		t.Append([]string{
			string(instance.ID),
			instance.Name,
			string(instance.Zone),
			instance.Type,
			instance.IPAddress,
			instance.IPv6Address,
			instance.State,
		})
	}
}

type instanceListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
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
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
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
	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {

		c := client.WithEndpoint(zone.APIEndpoint)
		instances, err := c.ListInstances(ctx)

		if err != nil {
			return err
		}

		for _, i := range instances.Instances {
			var instanceType *v3.InstanceType
			instanceTypeI, cached := instanceTypes.Load(i.InstanceType.ID)
			if cached {
				instanceType = instanceTypeI.(*v3.InstanceType)
			} else {
				instanceType, err = client.GetInstanceType(ctx, i.InstanceType.ID)
				if err != nil {
					return fmt.Errorf(
						"unable to retrieve Compute instance type %q: %w",
						i.InstanceType.ID,
						err)
				}
				instanceTypes.Store(i.InstanceType.ID, instanceType)
			}

			res <- instanceListItemOutput{
				ID:          i.ID,
				Name:        i.Name,
				Zone:        zone.Name,
				Type:        fmt.Sprintf("%s.%s", instanceType.Family, instanceType.Size),
				IPAddress:   utils.DefaultIP(&i.PublicIP, utils.EmptyIPAddressVisualization),
				IPv6Address: utils.DefaultIP(i.Ipv6Address, utils.EmptyIPAddressVisualization),
				State:       string(i.State),
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
