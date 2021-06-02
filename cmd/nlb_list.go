package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type nlbListItemOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Zone      string `json:"zone"`
	IPAddress string `json:"ip_address"`
}

type nlbListOutput []nlbListItemOutput

func (o *nlbListOutput) toJSON()  { outputJSON(o) }
func (o *nlbListOutput) toText()  { outputText(o) }
func (o *nlbListOutput) toTable() { outputTable(o) }

type nlbListCmd struct {
	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *nlbListCmd) cmdAliases() []string { return gListAlias }

func (c *nlbListCmd) cmdShort() string { return "List Network Load Balancers" }

func (c *nlbListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Network Load Balancers.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&nlbListItemOutput{}), ", "))
}

func (c *nlbListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	out := make(nlbListOutput, 0)
	res := make(chan nlbListItemOutput)
	defer close(res)

	go func() {
		for nlb := range res {
			out = append(out, nlb)
		}
	}()
	err := forEachZone(zones, func(zone string) error {
		list, err := cs.ListNetworkLoadBalancers(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Network Load Balancers in zone %s: %v", zone, err)
		}

		for _, nlb := range list {
			res <- nlbListItemOutput{
				ID:        nlb.ID,
				Name:      nlb.Name,
				Zone:      zone,
				IPAddress: nlb.IPAddress.String(),
			}
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbListCmd{}))
}
