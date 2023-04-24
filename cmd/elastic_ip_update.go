package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type elasticIPUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Description               string `cli-usage:"Elastic IP description"`
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
	ReverseDNS                string `cli-usage:"Reverse DNS Domain"`
}

func (c *elasticIPUpdateCmd) cmdAliases() []string { return nil }

func (c *elasticIPUpdateCmd) cmdShort() string {
	return "Update an Elastic IP"
}

func (c *elasticIPUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Compute instance Elastic IP.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updatedInstance, updatedRDNS bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	elasticIP, err := cs.FindElasticIP(ctx, c.Zone, c.ElasticIP)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		elasticIP.Description = &c.Description
		updatedInstance = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckMode)) {
		if elasticIP.Healthcheck == nil {
			elasticIP.Healthcheck = new(egoscale.ElasticIPHealthcheck)
		}
		elasticIP.Healthcheck.Mode = &c.HealthcheckMode
		updatedInstance = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.ReverseDNS)) {
		updatedRDNS = true
	}

	for _, flag := range []string{
		mustCLICommandFlagName(c, &c.HealthcheckInterval),
		mustCLICommandFlagName(c, &c.HealthcheckPort),
		mustCLICommandFlagName(c, &c.HealthcheckStrikesFail),
		mustCLICommandFlagName(c, &c.HealthcheckStrikesOK),
		mustCLICommandFlagName(c, &c.HealthcheckTLSSNI),
		mustCLICommandFlagName(c, &c.HealthcheckTLSSSkipVerify),
		mustCLICommandFlagName(c, &c.HealthcheckTimeout),
		mustCLICommandFlagName(c, &c.HealthcheckURI),
	} {
		if cmd.Flags().Changed(flag) && elasticIP.Healthcheck == nil {
			return fmt.Errorf("--%s cannot be used on a non-managed Elastic IP", flag)
		}
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckInterval); cmd.Flags().Changed(flag) {
		interval := time.Duration(c.HealthcheckInterval) * time.Second
		elasticIP.Healthcheck.Interval = &interval
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckPort); cmd.Flags().Changed(flag) {
		port := uint16(c.HealthcheckPort)
		elasticIP.Healthcheck.Port = &port
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckStrikesFail); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.StrikesFail = &c.HealthcheckStrikesFail
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckStrikesOK); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.StrikesOK = &c.HealthcheckStrikesOK
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckStrikesOK); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.StrikesOK = &c.HealthcheckStrikesOK
		updatedInstance = true
	}

	if elasticIP.Healthcheck != nil && *elasticIP.Healthcheck.Mode == "https" {
		if flag := mustCLICommandFlagName(c, &c.HealthcheckTLSSSkipVerify); cmd.Flags().Changed(flag) {
			elasticIP.Healthcheck.TLSSkipVerify = &c.HealthcheckTLSSSkipVerify
			updatedInstance = true
		}

		if flag := mustCLICommandFlagName(c, &c.HealthcheckTLSSNI); cmd.Flags().Changed(flag) {
			elasticIP.Healthcheck.TLSSNI = &c.HealthcheckTLSSNI
			updatedInstance = true
		}
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckTimeout); cmd.Flags().Changed(flag) {
		timeout := time.Duration(c.HealthcheckTimeout) * time.Second
		elasticIP.Healthcheck.Timeout = &timeout
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckURI); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.URI = &c.HealthcheckURI
		updatedInstance = true
	}

	if updatedInstance || updatedRDNS {
		decorateAsyncOperation(fmt.Sprintf("Updating Elastic IP %s...", c.ElasticIP), func() {
			if updatedInstance {
				if err = cs.UpdateElasticIP(ctx, c.Zone, elasticIP); err != nil {
					return
				}
			}

			if updatedRDNS {
				if c.ReverseDNS == "" {
					err = cs.DeleteElasticIPReverseDNS(ctx, c.Zone, *elasticIP.ID)
				} else {
					err = cs.UpdateElasticIPReverseDNS(ctx, c.Zone, *elasticIP.ID, c.ReverseDNS)
				}
			}
		})

		if err != nil {
			return err
		}
	}

	if !gQuiet {
		return (&elasticIPShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			ElasticIP:          *elasticIP.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
