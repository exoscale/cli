package vpc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcRouteListItemOutput struct {
	ID          v3.UUID `json:"id"`
	Kind        string  `json:"kind"`
	Destination string  `json:"destination"`
	Target      string  `json:"target"`
	Description string  `json:"description"`
}

type vpcRouteListOutput []vpcRouteListItemOutput

func (o *vpcRouteListOutput) ToJSON()  { output.JSON(o) }
func (o *vpcRouteListOutput) ToText()  { output.Text(o) }
func (o *vpcRouteListOutput) ToTable() { output.Table(o) }

type vpcRouteListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	VPC string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`

	Subnet string      `cli-usage:"only list routes of this Subnet (NAME|ID)"`
	Zone   v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcRouteListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *vpcRouteListCmd) CmdShort() string { return "List VPC routes" }

func (c *vpcRouteListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists the routes of a Virtual Private Cloud.

Without --subnet, the routes attached to the VPC are list.
With --subnet, all the routes impacting the Subnet are listed

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcRouteListItemOutput{}), ", "))
}

func (c *vpcRouteListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcRouteListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	out, err := c.list()
	if err != nil {
		return err
	}

	return c.OutputFunc(out, nil)
}

// list resolves the VPC (and Subnet, when --subnet is set) and returns the
// routes to display, ordered by kind.
func (c *vpcRouteListCmd) list() (*vpcRouteListOutput, error) {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return nil, err
	}

	vpcEntry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return nil, err
	}

	var routes []v3.ListRouteEntry

	if c.Subnet != "" {
		subnetEntry, err := FindSubnet(ctx, client, vpcEntry.ID, c.Subnet)
		if err != nil {
			return nil, err
		}

		resp, err := client.ListRoutes(ctx, vpcEntry.ID, subnetEntry.ID)
		if err != nil {
			return nil, err
		}
		routes = resp.Routes
	} else {
		resp, err := client.ListVpcRoutes(ctx, vpcEntry.ID)
		if err != nil {
			return nil, err
		}
		routes = resp.Routes
	}

	sortRoutesByKind(routes)

	out := make(vpcRouteListOutput, 0, len(routes))
	for _, r := range routes {
		out = append(out, vpcRouteListItemOutput{
			ID:          r.ID,
			Kind:        string(r.Kind),
			Destination: r.Destination,
			Target:      r.Target,
			Description: r.Description,
		})
	}

	return &out, nil
}

// sortRoutesByKind orders routes by kind, listing VPC routes before Subnet
// routes, then by destination so the output is stable.
func sortRoutesByKind(routes []v3.ListRouteEntry) {
	rank := func(k v3.ListRouteEntryKind) int {
		if k == v3.ListRouteEntryKindVpc {
			return 0
		}
		return 1
	}

	sort.SliceStable(routes, func(i, j int) bool {
		if ri, rj := rank(routes[i].Kind), rank(routes[j].Kind); ri != rj {
			return ri < rj
		}
		return routes[i].Destination < routes[j].Destination
	})
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcRouteCmd, &vpcRouteListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
