package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbCACertificateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"ca-certificate"`

	Zone string `cli-short:"z"`
}

func (c *dbCACertificateCmd) cmdAliases() []string { return nil }

func (c *dbCACertificateCmd) cmdShort() string { return "Retrieve the Database CA certificate" }

func (c *dbCACertificateCmd) cmdLong() string {
	return `This command retrieves the Exoscale organization-level CA certificate
required to access Database Services using a TLS connection.`
}

func (c *dbCACertificateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbCACertificateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	dbCACertificate, err := cs.GetDatabaseCACertificate(ctx, c.Zone)
	if err != nil {
		return err
	}

	_, _ = fmt.Print(dbCACertificate)

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbCmd, &dbCACertificateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
