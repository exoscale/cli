package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasCACertificateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

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
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasCACertificateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
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
	cobra.CheckErr(RegisterCLICommand(dbaasCmd, &dbaasCACertificateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
