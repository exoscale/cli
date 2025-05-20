package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPDeleteCmd) CmdAliases() []string { return GRemoveAlias }

func (c *elasticIPDeleteCmd) CmdShort() string {
	return "Delete an Elastic IP"
}

func (c *elasticIPDeleteCmd) CmdLong() string { return "" }

func (c *elasticIPDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	elasticIPs, err := client.ListElasticIPS(ctx)
	if err != nil {
		return err
	}

	eip, err := elasticIPs.FindElasticIP(c.ElasticIP)
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

	op, err := client.DeleteElasticIP(ctx, eip.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Elastic IP %s...", c.ElasticIP), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(elasticIPCmd, &elasticIPDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
