package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksClusterResetFields = []string{
	"description",
}

type sksUpdateCmd struct {
	_ bool `cli-cmd:"update"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description string   `cli-usage:"cluster description"`
	Name        string   `cli-usage:"cluster name"`
	ResetFields []string `cli-flag:"reset" cli-usage:"properties to reset to default value"`
	Zone        string   `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpdateCmd) cmdAliases() []string { return nil }

func (c *sksUpdateCmd) cmdShort() string { return "Update an SKS cluster" }

func (c *sksUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an SKS cluster.

Supported output template annotations: %s

Support values for --reset flag: %s`,
		strings.Join(outputterTemplateAnnotations(&sksShowOutput{}), ", "),
		strings.Join(sksClusterResetFields, ", "),
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
		cluster.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		cluster.Description = c.Description
		updated = true
	}

	decorateAsyncOperation(fmt.Sprintf("Updating SKS cluster %q...", c.Cluster), func() {
		if updated {
			err = cs.UpdateSKSCluster(ctx, c.Zone, cluster)
		}

		for _, f := range c.ResetFields {
			switch f {
			case "description":
				err = cluster.ResetField(ctx, &cluster.Description)
			}
			if err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showSKSCluster(c.Zone, cluster.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpdateCmd{}))
}
