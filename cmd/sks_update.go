package cmd

import (
	"errors"
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Cluster string `cli-arg:"#" cli-usage:"NAME|ID"`

	AutoUpgrade bool              `cli-usage:"enable automatic upgrading of the SKS cluster control plane Kubernetes version"`
	Description string            `cli-usage:"SKS cluster description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"SKS cluster label (format: key=value)"`
	Name        string            `cli-usage:"SKS cluster name"`
	Zone        string            `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksUpdateCmd) cmdAliases() []string { return nil }

func (c *sksUpdateCmd) cmdShort() string { return "Update an SKS cluster" }

func (c *sksUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an SKS cluster.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&sksShowOutput{}), ", "),
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
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.AutoUpgrade)) {
		cluster.AutoUpgrade = &c.AutoUpgrade
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		cluster.Labels = &c.Labels
		updated = true
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
		return (&sksShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Cluster:            *cluster.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
