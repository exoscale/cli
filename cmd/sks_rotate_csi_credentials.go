package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksRotateCSICredentialsCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"rotate-csi-credentials"`

	Cluster string      `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Zone    v3.ZoneName `cli-flag:"zone" cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksRotateCSICredentialsCmd) CmdAliases() []string { return nil }

func (c *sksRotateCSICredentialsCmd) CmdShort() string {
	return "Rotate the Exoscale Container Storage Interface IAM credentials for an SKS cluster"
}

func (c *sksRotateCSICredentialsCmd) CmdLong() string {
	return `This command rotates the Exoscale IAM credentials managed by the SKS control
plane for the Kubernetes Exoscale Container Storage Interface.
`
}

func (c *sksRotateCSICredentialsCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksRotateCSICredentialsCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	op, err := client.RotateSKSCsiCredentials(ctx, cluster.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Rotating SKS cluster %q Exoscale CSI credentials...", c.Cluster), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(RegisterCLICommand(sksCmd, &sksRotateCSICredentialsCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
