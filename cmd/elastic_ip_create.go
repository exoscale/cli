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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Description               string      `cli-usage:"Elastic IP description"`
	IPv6                      bool        `cli-flag:"ipv6" cli-usage:"create Elastic IPv6 prefix"`
	HealthcheckInterval       int64       `cli-usage:"managed Elastic IP health checking interval in seconds"`
	HealthcheckMode           string      `cli-usage:"managed Elastic IP health checking mode (tcp|http|https)"`
	HealthcheckPort           int64       `cli-usage:"managed Elastic IP health checking port"`
	HealthcheckStrikesFail    int64       `cli-usage:"number of failed attempts before considering a managed Elastic IP health check unhealthy"`
	HealthcheckStrikesOK      int64       `cli-usage:"number of successful attempts before considering a managed Elastic IP health check healthy"`
	HealthcheckTLSSNI         string      `cli-flag:"healthcheck-tls-sni" cli-usage:"managed Elastic IP health checking server name to present with SNI in https mode"`
	HealthcheckTLSSSkipVerify bool        `cli-flag:"healthcheck-tls-skip-verify" cli-usage:"disable TLS certificate verification for managed Elastic IP health checking in https mode"`
	HealthcheckTimeout        int64       `cli-usage:"managed Elastic IP health checking timeout in seconds"`
	HealthcheckURI            string      `cli-usage:"managed Elastic IP health checking URI (required in http(s) mode)"`
	Zone                      v3.ZoneName `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *elasticIPCreateCmd) cmdShort() string {
	return "Create an Elastic IP"
}

func (c *elasticIPCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Elastic IP.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	var healthcheck *v3.ElasticIPHealthcheck
	if c.HealthcheckMode != "" {
		healthcheck = &v3.ElasticIPHealthcheck{
			Interval:    c.HealthcheckInterval,
			Mode:        v3.ElasticIPHealthcheckMode(c.HealthcheckMode),
			Port:        c.HealthcheckPort,
			StrikesFail: c.HealthcheckStrikesFail,
			StrikesOk:   c.HealthcheckStrikesOK,
			Timeout:     c.HealthcheckTimeout,
			URI: func() (v string) {
				if strings.HasPrefix(c.HealthcheckMode, "http") {
					v = c.HealthcheckURI
				}
				return
			}(),
		}

		if c.HealthcheckMode == "https" {
			healthcheck.TlsSkipVerify = &c.HealthcheckTLSSSkipVerify
			healthcheck.TlsSNI = c.HealthcheckTLSSNI
		}
	}

	createElasticIPRequest := v3.CreateElasticIPRequest{
		Description: c.Description,
		Healthcheck: healthcheck,
	}

	if c.IPv6 {
		createElasticIPRequest.Addressfamily = v3.CreateElasticIPRequestAddressfamilyInet6
	}

	err = decorateAsyncOperations("Creating Elastic IP...", func() error {
		op, err := client.CreateElasticIP(ctx, createElasticIPRequest)
		if err != nil {
			return fmt.Errorf("exoscale: error while creating Elastic IP: %w", err)
		}

		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting for Elastic IP creation: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return (&elasticIPShowCmd{
		cliCommandSettings: c.cliCommandSettings,
		// Comment for reviewer:
		// Is there a way to get the created Elastic IP address or UUID ???
		// Listing all of them and comparing Addressfamily, Description,... doesn't really garantee uniquess
		// TODO: Remove comment before merging
		ElasticIP: "",
		Zone:      c.Zone,
	}).cmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		HealthcheckInterval:    10,
		HealthcheckStrikesFail: 2,
		HealthcheckStrikesOK:   3,
		HealthcheckTimeout:     3,
	}))
}
