package sks

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksRotateOperatorsCACmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"rotate-operators-ca"`

	Cluster string      `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Zone    v3.ZoneName `cli-flag:"zone" cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksRotateOperatorsCACmd) CmdAliases() []string { return nil }

func (c *sksRotateOperatorsCACmd) CmdShort() string {
	return "Rotate the Exoscale Operators Certificate Authority for an SKS cluster"
}

func (c *sksRotateOperatorsCACmd) CmdLong() string {
	return `This command rotates the Exoscale certificate authority (CA) used by Kubernetes operators within the SKS control plane, ensuring secure communication and certificate management for cluster operations.
`
}

func (c *sksRotateOperatorsCACmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksRotateOperatorsCACmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	op, err := client.RotateSKSOperatorsCA(ctx, cluster.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Rotating SKS Cluster Operators CA %q...", c.Cluster), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksRotateOperatorsCACmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
