package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/output"
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

func (o *deployTargetListOutput) toJSON()  { output.JSON(o) }
func (o *deployTargetListOutput) toText()  { output.Text(o) }
func (o *deployTargetListOutput) toTable() { output.Table(o) }

type deployTargetListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *deployTargetListCmd) cmdAliases() []string { return nil }

func (c *deployTargetListCmd) cmdShort() string { return "List Deploy Targets" }

func (c *deployTargetListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists existing Deploy Targets.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&deployTargetListOutput{}), ", "))
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

	out := make(deployTargetListOutput, 0)
	res := make(chan deployTargetListItemOutput)
	done := make(chan struct{})

	go func() {
		for dt := range res {
			out = append(out, dt)
		}
		done <- struct{}{}
	}()
	err := forEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		list, err := cs.ListDeployTargets(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Deploy Targets in zone %s: %w", zone, err)
		}

		for _, dt := range list {
			res <- deployTargetListItemOutput{
				ID:   *dt.ID,
				Name: *dt.Name,
				Type: *dt.Type,
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
	cobra.CheckErr(registerCLICommand(deployTargetCmd, &deployTargetListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
