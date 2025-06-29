package dbaas

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c dbaasDatabaseDeleteCmd) deletePg(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	s, err := client.GetDBAASServicePG(ctx, c.Name)
	if err != nil {
		return err
	}

	dbFound := false
	for _, db := range s.Databases {
		if db == v3.DBAASDatabaseName(c.Database) {
			dbFound = true
			break
		}
	}

	if !dbFound {
		return fmt.Errorf("database %q not found for service %q", c.Database, c.Name)
	}
	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to delete database %q", c.Database)) {
			return nil
		}
	}

	op, err := client.DeleteDBAASPGDatabase(ctx, c.Name, c.Database)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting DBaaS database %q", c.Database), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServicePG(ctx))
	}

	return err
}
