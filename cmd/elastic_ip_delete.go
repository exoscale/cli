package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *elasticIPDeleteCmd) cmdShort() string {
	return "Delete an Elastic IP"
}

func (c *elasticIPDeleteCmd) cmdLong() string { return "" }

func (c *elasticIPDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	elasticIPResp, err := client.ListElasticIPS(ctx)
	if err != nil {
		return err
	}

	elasticIP, err := elasticIPResp.FindElasticIP(c.ElasticIP)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Elastic IP %s?", c.ElasticIP)) {
			return nil
		}
	}

	return decorateAsyncOperations(fmt.Sprintf("Deleting Elastic IP %s...", c.ElasticIP), func() error {
		op, err := client.DeleteElasticIP(ctx, elasticIP.ID)
		if err != nil {
			return fmt.Errorf("exoscale: error while deleting Elastic IP: %w", err)
		}

		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		return err
	})
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
