package dbaas

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasReadReplicaPromoteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"promote"`

	Name  string `cli-arg:"#" cli-usage:"REPLICA-NAME"`
	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasReadReplicaPromoteCmd) CmdAliases() []string { return nil }

func (c *dbaasReadReplicaPromoteCmd) CmdShort() string {
	return "Promote a Database Service read replica to a standalone primary"
}

func (c *dbaasReadReplicaPromoteCmd) CmdLong() string {
	return "Promote a read replica to a standalone primary Database Service. This breaks the replication link."
}

func (c *dbaasReadReplicaPromoteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasReadReplicaPromoteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to promote read replica %q? This will break the replication link.", c.Name)) {
			return nil
		}
	}

	client, err := dbaasReadReplicaClientForZone(ctx, c.Zone)
	if err != nil {
		return err
	}

	databaseService, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !dbaasReadReplicaSupportedServiceType(string(databaseService.Type)) {
		return fmt.Errorf("read replicas are not supported for Database Service type %q", databaseService.Type)
	}

	replicaIntegration := dbaasGetReadReplicaIntegrationForReplica(databaseService)
	if replicaIntegration == nil {
		return fmt.Errorf("%q is not a read replica", c.Name)
	}

	op, err := client.DeleteDBAASIntegration(ctx, replicaIntegration.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Promoting read replica %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		showCmd := &dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}

		switch databaseService.Type {
		case "pg":
			return c.OutputFunc(showCmd.showDatabaseServicePG(ctx))
		case "mysql":
			return c.OutputFunc(showCmd.showDatabaseServiceMysql(ctx))
		default:
			return nil
		}
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasReadReplicaCmd, &dbaasReadReplicaPromoteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
