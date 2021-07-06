package cmd

import (
	"fmt"
	"os"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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

func (o *instanceListOutput) toJSON()  { outputJSON(o) }
func (o *instanceListOutput) toText()  { outputText(o) }
func (o *instanceListOutput) toTable() { outputTable(o) }

type instanceListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *instanceListCmd) cmdAliases() []string { return gListAlias }

func (c *instanceListCmd) cmdShort() string { return "List Compute instances" }

func (c *instanceListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Compute instances.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceListItemOutput{}), ", "))
}

func (c *instanceListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	out := make(instanceListOutput, 0)
	res := make(chan instanceListItemOutput)
	defer close(res)

	instanceTypes := make(map[string]*exov2.InstanceType) // For caching

	go func() {
		for instance := range res {
			out = append(out, instance)
		}
	}()
	err := forEachZone(zones, func(zone string) error {
		list, err := cs.ListInstances(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Compute instances in zone %s: %v", zone, err)
		}

		for _, i := range list {
			instanceType, cached := instanceTypes[*i.InstanceTypeID]
			if !cached {
				instanceType, err = cs.GetInstanceType(ctx, zone, *i.InstanceTypeID)
				if err != nil {
					return fmt.Errorf(
						"unable to retrieve Compute instance type %q: %s",
						*i.InstanceTypeID,
						err)
				}
				instanceTypes[*i.InstanceTypeID] = instanceType
			}

			res <- instanceListItemOutput{
				ID:        *i.ID,
				Name:      *i.Name,
				Zone:      zone,
				Type:      fmt.Sprintf("%s.%s", *instanceType.Family, *instanceType.Size),
				IPAddress: i.PublicIPAddress.String(),
				State:     *i.State,
			}
		}

		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceCmd, &instanceListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
