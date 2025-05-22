package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPCreateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Description               string `cli-usage:"Elastic IP description"`
	IPv6                      bool   `cli-flag:"ipv6" cli-usage:"create Elastic IPv6 prefix"`
	HealthcheckInterval       int64  `cli-usage:"managed Elastic IP health checking interval in seconds"`
	HealthcheckMode           string `cli-usage:"managed Elastic IP health checking mode (tcp|http|https)"`
	HealthcheckPort           int64  `cli-usage:"managed Elastic IP health checking port"`
	HealthcheckStrikesFail    int64  `cli-usage:"number of failed attempts before considering a managed Elastic IP health check unhealthy"`
	HealthcheckStrikesOK      int64  `cli-usage:"number of successful attempts before considering a managed Elastic IP health check healthy"`
	HealthcheckTLSSNI         string `cli-flag:"healthcheck-tls-sni" cli-usage:"managed Elastic IP health checking server name to present with SNI in https mode"`
	HealthcheckTLSSSkipVerify bool   `cli-flag:"healthcheck-tls-skip-verify" cli-usage:"disable TLS certificate verification for managed Elastic IP health checking in https mode"`
	HealthcheckTimeout        int64  `cli-usage:"managed Elastic IP health checking timeout in seconds"`
	HealthcheckURI            string `cli-usage:"managed Elastic IP health checking URI (required in http(s) mode)"`
	Zone                      string `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPCreateCmd) CmdAliases() []string { return GCreateAlias }

func (c *elasticIPCreateCmd) CmdShort() string {
	return "Create an Elastic IP"
}

func (c *elasticIPCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Elastic IP.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	var healthcheck *v3.ElasticIPHealthcheck
	if c.HealthcheckMode != "" {

		healthcheck = &v3.ElasticIPHealthcheck{
			Interval:    c.HealthcheckInterval,
			StrikesFail: c.HealthcheckStrikesFail,
			StrikesOk:   c.HealthcheckStrikesOK,
			Timeout:     c.HealthcheckTimeout,
			Mode:        v3.ElasticIPHealthcheckMode(c.HealthcheckMode),
			Port:        c.HealthcheckPort,
		}
		if strings.HasPrefix(c.HealthcheckMode, "http") {
			healthcheck.URI = c.HealthcheckURI
		}

		if c.HealthcheckMode == "https" {
			healthcheck.TlsSkipVerify = &c.HealthcheckTLSSSkipVerify
			if c.HealthcheckTLSSNI != "" {
				healthcheck.TlsSNI = c.HealthcheckTLSSNI
			}
		}
	}

	elasticIP := v3.CreateElasticIPRequest{
		Healthcheck: healthcheck,
		Description: c.Description,
	}

	if c.IPv6 {
		elasticIP.Addressfamily = "inetv6"
	}

	op, err := client.CreateElasticIP(ctx, elasticIP)
	if err != nil {
		return err
	}

	decorateAsyncOperation("Creating Elastic IP...", func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	return (&elasticIPShowCmd{
		CliCommandSettings: c.CliCommandSettings,
		ElasticIP:          op.Reference.ID.String(),
		Zone:               c.Zone,
	}).CmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(elasticIPCmd, &elasticIPCreateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),

		HealthcheckInterval:    10,
		HealthcheckStrikesFail: 2,
		HealthcheckStrikesOK:   3,
		HealthcheckTimeout:     3,
	}))
}
