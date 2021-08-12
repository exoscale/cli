package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksRotateCCMCredentialsCmd struct {
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
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	decorateAsyncOperation(
		fmt.Sprintf("Rotating SKS cluster %q Exoscale CCM credentials...", c.Cluster),
		func() {
			err = cs.RotateSKSClusterCCMCredentials(ctx, c.Zone, cluster)
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksRotateCCMCredentialsCmd{}))
}
