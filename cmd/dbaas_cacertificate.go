package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
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
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasCACertificateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
