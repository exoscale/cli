package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c dbaasDatabaseCreateCmd) createMysql(cmd *cobra.Command, _ []string) error {

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	s, err := client.GetDBAASServiceMysql(ctx, c.Name)
	if err != nil {
		return err
	}

	if len(s.Databases) == 0 {
		return fmt.Errorf("service %q is not ready for database creation", c.Name)
	}

	req := v3.CreateDBAASMysqlDatabaseRequest{
		DatabaseName: v3.DBAASDatabaseName(c.Database),
	}

	op, err := client.CreateDBAASMysqlDatabase(ctx, c.Name, req)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating DBaaS database %q", c.Database), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceMysql(ctx))
	}

	return err
}
