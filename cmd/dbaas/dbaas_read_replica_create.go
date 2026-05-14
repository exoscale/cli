package dbaas

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasReadReplicaCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"READ-REPLICA-NAME"`

	Plan                  string `cli-flag:"plan" cli-usage:"subscription plan"`
	ReplicaZone           string `cli-flag:"replica-zone" cli-short:"z" cli-usage:"zone where the replica will be created"`
	SourceService         string `cli-flag:"source-service" cli-usage:"name of the primary service"`
	TerminationProtection bool   `cli-usage:"enable Database Service termination protection; set --termination-protection=false to disable"`
}

func (c *dbaasReadReplicaCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *dbaasReadReplicaCreateCmd) CmdShort() string {
	return "Create a Database Service read replica"
}

func (c *dbaasReadReplicaCreateCmd) CmdLong() string {
	return "Create a read replica for an existing PostgreSQL or MySQL/MariaDB primary Database Service. The replica can be created in a different zone than the primary."
}

func (c *dbaasReadReplicaCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)

	if c.SourceService == "" {
		return fmt.Errorf("--source-service is required")
	}
	if c.ReplicaZone == "" {
		return fmt.Errorf("--replica-zone is required")
	}
	if c.Plan == "" {
		return fmt.Errorf("--plan is required")
	}

}

func (c *dbaasReadReplicaCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	sourceService, err := dbaasFindServiceByNameAllZones(ctx, c.SourceService)
	if err != nil {
		return err
	}

	client, err := dbaasReadReplicaClientForZone(ctx, c.ReplicaZone)
	if err != nil {
		return err
	}

	switch sourceService.Service.Type {
	case "pg":
		databaseService := v3.CreateDBAASServicePGRequest{
			Plan:                  c.Plan,
			TerminationProtection: &c.TerminationProtection,
			Integrations: []v3.CreateDBAASServicePGRequestIntegrations{
				{
					Type:          v3.CreateDBAASServicePGRequestIntegrationsTypeReadReplica,
					SourceService: v3.DBAASServiceName(c.SourceService),
				},
			},
		}

		op, err := client.CreateDBAASServicePG(ctx, c.Name, databaseService)
		if err != nil {
			return err
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Creating read replica %q...", c.Name), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	case "mysql":
		databaseService := v3.CreateDBAASServiceMysqlRequest{
			Plan:                  c.Plan,
			TerminationProtection: &c.TerminationProtection,
			Integrations: []v3.CreateDBAASServiceMysqlRequestIntegrations{
				{
					Type:          v3.CreateDBAASServiceMysqlRequestIntegrationsTypeReadReplica,
					SourceService: v3.DBAASServiceName(c.SourceService),
				},
			},
		}

		op, err := client.CreateDBAASServiceMysql(ctx, c.Name, databaseService)
		if err != nil {
			return err
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Creating read replica %q...", c.Name), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("read replicas are not supported for Database Service type %q", sourceService.Service.Type)
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasReadReplicaShowCmd{
			Name: c.Name,
			Zone: c.ReplicaZone,
		}).showReadReplica(ctx))
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasReadReplicaCmd, &dbaasReadReplicaCreateCmd{
		CliCommandSettings:    exocmd.DefaultCLICmdSettings(),
		TerminationProtection: true,
	}))
}
