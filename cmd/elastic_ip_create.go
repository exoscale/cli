package cmd

import (
	"fmt"
	"strings"
	"time"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type elasticIPCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Description               string `cli-usage:"Elastic IP description"`
	IPv6                      bool   `cli-flag:"ipv6" cli-usage:"create IPv6 Elastic IP address instead of IPv4"`
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

func (c *elasticIPCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *elasticIPCreateCmd) cmdShort() string {
	return "Create an Elastic IP"
}

func (c *elasticIPCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Elastic IP.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	var healthcheck *egoscale.ElasticIPHealthcheck
	if c.HealthcheckMode != "" {
		port := uint16(c.HealthcheckPort)
		interval := time.Duration(c.HealthcheckInterval) * time.Second
		timeout := time.Duration(c.HealthcheckTimeout) * time.Second

		healthcheck = &egoscale.ElasticIPHealthcheck{
			Interval:    &interval,
			Mode:        &c.HealthcheckMode,
			Port:        &port,
			StrikesFail: &c.HealthcheckStrikesFail,
			StrikesOK:   &c.HealthcheckStrikesOK,
			Timeout:     &timeout,
			URI: func() (v *string) {
				if strings.HasPrefix(c.HealthcheckMode, "http") {
					v = &c.HealthcheckURI
				}
				return
			}(),
		}

		if c.HealthcheckMode == "https" {
			healthcheck.TLSSkipVerify = &c.HealthcheckTLSSSkipVerify
			healthcheck.TLSSNI = nonEmptyStringPtr(c.HealthcheckTLSSNI)
		}
	}

	elasticIP := &egoscale.ElasticIP{
		Description: nonEmptyStringPtr(c.Description),
		Healthcheck: healthcheck,
	}

	if c.IPv6 {
		elasticIP.AddressFamily = nonEmptyStringPtr("inet6")
	}

	var err error
	decorateAsyncOperation("Creating Elastic IP...", func() {
		elasticIP, err = cs.CreateElasticIP(ctx, c.Zone, elasticIP)
	})
	if err != nil {
		return err
	}

	return (&elasticIPShowCmd{
		cliCommandSettings: c.cliCommandSettings,
		ElasticIP:          *elasticIP.ID,
		Zone:               c.Zone,
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
