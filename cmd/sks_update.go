package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksUpdateCmd struct {
	_ bool `cli-cmd:"update"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description string `cli-usage:"cluster description"`
	Name        string `cli-usage:"cluster name"`
	Zone        string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpdateCmd) cmdAliases() []string { return nil }

func (c *sksUpdateCmd) cmdShort() string { return "Update an SKS cluster" }

func (c *sksUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an SKS cluster.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksShowOutput{}), ", "),
	)
}

func (c *sksUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		cluster.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		cluster.Description = &c.Description
		updated = true
	}

	if updated {
		decorateAsyncOperation(fmt.Sprintf("Updating SKS cluster %q...", c.Cluster), func() {
			err = cs.UpdateSKSCluster(ctx, c.Zone, cluster)
		})
		if err != nil {
			return err
		}
	}

	if !gQuiet {
		return output(showSKSCluster(c.Zone, *cluster.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpdateCmd{}))
}
