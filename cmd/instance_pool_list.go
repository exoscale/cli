package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolListItemOutput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Size  int64  `json:"size"`
	State string `json:"state"`
}

type instancePoolListOutput []instancePoolListItemOutput

func (o *instancePoolListOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolListOutput) toText()  { outputText(o) }
func (o *instancePoolListOutput) toTable() { outputTable(o) }

type instancePoolListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *instancePoolListCmd) cmdAliases() []string { return gListAlias }

func (c *instancePoolListCmd) cmdShort() string { return "List Instance Pools" }

func (c *instancePoolListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Instance Pools.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolListItemOutput{}), ", "))
}

func (c *instancePoolListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	out := make(instancePoolListOutput, 0)
	res := make(chan instancePoolListItemOutput)
	defer close(res)

	go func() {
		for instancePool := range res {
			out = append(out, instancePool)
		}
	}()
	err := forEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		list, err := cs.ListInstancePools(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Instance Pools in zone %s: %v", zone, err)
		}

		for _, i := range list {
			res <- instancePoolListItemOutput{
				ID:    *i.ID,
				Name:  *i.Name,
				Zone:  zone,
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

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedInstancePoolCmd, &instancePoolListCmd{}))
}
