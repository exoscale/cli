package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type privateNetworkListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Zone string `json:"zone"`
}

type privateNetworkListOutput []privateNetworkListItemOutput

func (o *privateNetworkListOutput) toJSON()  { output.JSON(o) }
func (o *privateNetworkListOutput) toText()  { output.Text(o) }
func (o *privateNetworkListOutput) toTable() { output.Table(o) }

type privateNetworkListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *privateNetworkListCmd) cmdAliases() []string { return gListAlias }

func (c *privateNetworkListCmd) cmdShort() string { return "List Private Networks" }

func (c *privateNetworkListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Private Networks.

Supported output template annotations: %s`,
		strings.Join(output.output.OutputterTemplateAnnotations(&privateNetworkListItemOutput{}), ", "))
}

func (c *privateNetworkListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	out := make(privateNetworkListOutput, 0)
	res := make(chan privateNetworkListItemOutput)
	done := make(chan struct{})

	go func() {
		for nlb := range res {
			out = append(out, nlb)
		}
		done <- struct{}{}
	}()
	err := forEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		list, err := cs.ListPrivateNetworks(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Private Networks in zone %s: %w", zone, err)
		}

		for _, p := range list {
			res <- privateNetworkListItemOutput{
				ID:   *p.ID,
				Name: *p.Name,
				Zone: zone,
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

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
