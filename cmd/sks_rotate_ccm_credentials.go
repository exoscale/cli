package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type sksRotateCCMCredentialsCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"rotate-ccm-credentials"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`

	Zone string `cli-flag:"zone" cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksRotateCCMCredentialsCmd) cmdAliases() []string { return nil }

func (c *sksRotateCCMCredentialsCmd) cmdShort() string {
	return "Rotate the Exoscale Cloud Controller IAM credentials for an SKS cluster"
}

func (c *sksRotateCCMCredentialsCmd) cmdLong() string {
	return `This command rotates the Exoscale IAM credentials managed by the SKS control
plane for the Kubernetes Exoscale Cloud Controller Manager.
`
}

func (c *sksRotateCCMCredentialsCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksRotateCCMCredentialsCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	cluster, err := globalstate.EgoscaleClient.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	decorateAsyncOperation(
		fmt.Sprintf("Rotating SKS cluster %q Exoscale CCM credentials...", c.Cluster),
		func() {
			err = globalstate.EgoscaleClient.RotateSKSClusterCCMCredentials(ctx, c.Zone, cluster)
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksRotateCCMCredentialsCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksRotateCCMCredentialsCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
