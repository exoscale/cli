package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type deployTargetListItemOutput struct {
	Zone string `json:"zone"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type deployTargetListOutput []deployTargetListItemOutput

func (o *deployTargetListOutput) toJSON()  { outputJSON(o) }
func (o *deployTargetListOutput) toText()  { outputText(o) }
func (o *deployTargetListOutput) toTable() { outputTable(o) }

type deployTargetListCmd struct {
	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *deployTargetListCmd) cmdAliases() []string { return nil }

func (c *deployTargetListCmd) cmdShort() string { return "List Deploy Targets" }

func (c *deployTargetListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists existing Deploy Targets.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&vmListOutput{}), ", "))
}

func (c *deployTargetListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *deployTargetListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zones[0]))

	out := make(deployTargetListOutput, 0)
	res := make(chan deployTargetListItemOutput)
	defer close(res)

	go func() {
		for dt := range res {
			out = append(out, dt)
		}
	}()
	err := forEachZone(zones, func(zone string) error {
		list, err := cs.ListDeployTargets(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Deploy Targets in zone %s: %v", zone, err)
		}

		for _, dt := range list {
			res <- deployTargetListItemOutput{
				ID:   dt.ID,
				Name: dt.Name,
				Type: dt.Type,
				Zone: zone,
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
	cobra.CheckErr(registerCLICommand(deployTargetCmd, &deployTargetListCmd{}))
}
