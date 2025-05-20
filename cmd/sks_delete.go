package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force           bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	DeleteNodepools bool        `cli-flag:"nodepools" cli-short:"n" cli-usage:"delete existing Nodepools before deleting the SKS cluster"`
	Zone            v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksDeleteCmd) CmdAliases() []string { return GRemoveAlias }

func (c *sksDeleteCmd) CmdShort() string { return "Delete an SKS cluster" }

func (c *sksDeleteCmd) CmdLong() string { return "" }

func (c *sksDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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

	if len(cluster.Nodepools) > 0 {
		nodepoolsRemaining := len(cluster.Nodepools)

		if c.DeleteNodepools {
			for _, nodepool := range cluster.Nodepools {
				nodepool := nodepool

				if !c.Force {
					if !askQuestion(fmt.Sprintf(
						"Are you sure you want to delete Nodepool %q?",
						nodepool.Name),
					) {
						continue
					}
				}

				op, err := client.DeleteSKSNodepool(ctx, cluster.ID, nodepool.ID)
				if err != nil {
					return err
				}

				decorateAsyncOperation(fmt.Sprintf("Deleting Nodepool %q...", nodepool.Name), func() {
					_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
				})
				if err != nil {
					return err
				}
				nodepoolsRemaining--
			}
		}

		// It's not possible to delete an SKS cluster that still has Nodepools, no need to go further.
		if nodepoolsRemaining > 0 {
			return errors.New("impossible to delete the SKS cluster: Nodepools still present")
		}
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete SKS cluster %q?", cluster.Name)) {
			return nil
		}
	}

	op, err := client.DeleteSKSCluster(ctx, cluster.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting SKS cluster %q...", cluster.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(sksCmd, &sksDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
