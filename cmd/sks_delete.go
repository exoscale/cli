package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type sksDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force           bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	DeleteNodepools bool   `cli-flag:"nodepools" cli-short:"n" cli-usage:"delete existing Nodepools before deleting the SKS cluster"`
	Zone            string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *sksDeleteCmd) cmdShort() string { return "Delete an SKS cluster" }

func (c *sksDeleteCmd) cmdLong() string { return "" }

func (c *sksDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	cluster, err := globalstate.EgoscaleClient.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
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
						*nodepool.Name),
					) {
						continue
					}
				}

				decorateAsyncOperation(fmt.Sprintf("Deleting Nodepool %q...", *nodepool.Name), func() {
					err = globalstate.EgoscaleClient.DeleteSKSNodepool(ctx, c.Zone, cluster, nodepool)
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
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete SKS cluster %q?", *cluster.Name)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting SKS cluster %q...", *cluster.Name), func() {
		err = globalstate.EgoscaleClient.DeleteSKSCluster(ctx, c.Zone, cluster)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
