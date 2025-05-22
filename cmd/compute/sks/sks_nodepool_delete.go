package sks

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksNodepoolDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Cluster  string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Nodepool string `cli-arg:"#" cli-usage:"NODEPOOL-NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksNodepoolDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *sksNodepoolDeleteCmd) CmdShort() string { return "Delete an SKS cluster Nodepool" }

func (c *sksNodepoolDeleteCmd) CmdLong() string { return "" }

func (c *sksNodepoolDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if !c.Force {
		if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf("Are you sure you want to delete Nodepool %q?", c.Nodepool)) {
			return nil
		}
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := resp.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	for _, nodepool := range cluster.Nodepools {
		if nodepool.ID.String() == c.Nodepool || nodepool.Name == c.Nodepool {
			nodepool := nodepool

			op, err := client.DeleteSKSNodepool(ctx, cluster.ID, nodepool.ID)
			if err != nil {
				return err
			}
			utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Nodepool %q...", nodepool.Name), func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})
			if err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("nodepool not found") // nolint:stylecheck
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksNodepoolCmd, &sksNodepoolDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
