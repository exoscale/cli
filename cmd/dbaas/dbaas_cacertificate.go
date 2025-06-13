package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasCACertificateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"ca-certificate"`

	Zone string `cli-short:"z"`
}

func (c *dbaasCACertificateCmd) CmdAliases() []string { return nil }

func (c *dbaasCACertificateCmd) CmdShort() string { return "Retrieve the Database CA certificate" }

func (c *dbaasCACertificateCmd) CmdLong() string {
	return `This command retrieves the Exoscale organization-level CA certificate
required to access Database Services using a TLS connection.`
}

func (c *dbaasCACertificateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

<<<<<<< Updated upstream:cmd/dbaas_cacertificate.go
func (c *dbaasCACertificateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
=======
func (c *dbaasCACertificateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	caCertificate, err := globalstate.EgoscaleClient.GetDatabaseCACertificate(ctx, c.Zone)
>>>>>>> Stashed changes:cmd/dbaas/dbaas_cacertificate.go
	if err != nil {
		return err
	}

	caCertificate, err := client.GetDBAASCACertificate(ctx)
	if err != nil {
		return err
	}
	_, _ = fmt.Print(caCertificate.Certificate)

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasCACertificateCmd{
		cliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
