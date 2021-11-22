package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasCACertificateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"ca-certificate"`

	Zone string `cli-short:"z"`
}

func (c *dbaasCACertificateCmd) cmdAliases() []string { return nil }

func (c *dbaasCACertificateCmd) cmdShort() string { return "Retrieve the Database CA certificate" }

func (c *dbaasCACertificateCmd) cmdLong() string {
	return `This command retrieves the Exoscale organization-level CA certificate
required to access Database Services using a TLS connection.`
}

func (c *dbaasCACertificateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasCACertificateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	caCertificate, err := cs.GetDatabaseCACertificate(ctx, c.Zone)
	if err != nil {
		return err
	}

	_, _ = fmt.Print(caCertificate)

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasCACertificateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
