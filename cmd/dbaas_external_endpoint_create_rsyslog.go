package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

func (c *dbaasExternalEndpointCreateCmd) createRsyslog(cmd *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	rsyslogRequestPayload := v3.DBAASEndpointRsyslogInputCreate{
		Settings: &v3.DBAASEndpointRsyslogInputCreateSettings{},
	}

	if c.RsyslogCA != "" {
		rsyslogRequestPayload.Settings.CA = c.RsyslogCA
	}
	if c.RsyslogCert != "" {
		rsyslogRequestPayload.Settings.Cert = c.RsyslogCert
	}
	if c.RsyslogFormat != "" {
		rsyslogRequestPayload.Settings.Format = v3.EnumRsyslogFormat(c.RsyslogFormat)
	}
	if c.RsyslogLogline != "" {
		rsyslogRequestPayload.Settings.Logline = c.RsyslogLogline
	}
	if c.RsyslogKey != "" {
		rsyslogRequestPayload.Settings.Key = c.RsyslogKey
	}
	if c.RsyslogPort != 0 {
		rsyslogRequestPayload.Settings.Port = c.RsyslogPort
	}
	if c.RsyslogMaxMessageSize != 0 {
		rsyslogRequestPayload.Settings.MaxMessageSize = c.RsyslogMaxMessageSize
	}
	if c.RsyslogSD != "" {
		rsyslogRequestPayload.Settings.SD = c.RsyslogSD
	}
	if c.RsyslogServer != "" {
		rsyslogRequestPayload.Settings.Server = c.RsyslogServer
	}
	if cmd.Flags().Changed("rsyslog-tls") {
		rsyslogRequestPayload.Settings.Tls = v3.Bool(c.RsyslogTls)
	}

	op, err := client.CreateDBAASExternalEndpointRsyslog(ctx, c.Name, rsyslogRequestPayload)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS Rsyslog external Endpoint %q", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	endpointID := op.Reference.ID.String()
	if !globalstate.Quiet {
		return (&dbaasExternalEndpointShowCmd{
			CliCommandSettings: DefaultCLICmdSettings(),
			EndpointID:         endpointID,
			Type:               "rsyslog",
		}).CmdRun(nil, nil)
	}
	return nil
}
