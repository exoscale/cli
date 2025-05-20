package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c dbaasDatabaseCreateCmd) createPg(cmd *cobra.Command, _ []string) error {
	ctx := GContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
	s, err := client.GetDBAASServicePG(ctx, c.Name)
	if err != nil {
		return err
	}

	if len(s.Databases) == 0 {
		return fmt.Errorf("service %q is not ready for database creation", c.Name)
	}

	req := v3.CreateDBAASPGDatabaseRequest{
		DatabaseName: v3.DBAASDatabaseName(c.Database),
	}

	if c.PgLcCollate != "" {
		req.LCCollate = c.PgLcCollate
	}

	if c.PgLcCtype != "" {
		req.LCCtype = c.PgLcCtype
	}

	op, err := client.CreateDBAASPGDatabase(ctx, c.Name, req)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS database %q", c.Database), func() {
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
