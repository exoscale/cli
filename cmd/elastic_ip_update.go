package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Description               string      `cli-usage:"Elastic IP description"`
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
	ReverseDNS                string      `cli-usage:"Reverse DNS Domain"`
}

func (c *elasticIPUpdateCmd) cmdAliases() []string { return nil }

func (c *elasticIPUpdateCmd) cmdShort() string {
	return "Update an Elastic IP"
}

func (c *elasticIPUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Compute instance Elastic IP.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updatedInstance, updatedRDNS bool

	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	elasticIPResp, err := client.ListElasticIPS(ctx)
	if err != nil {
		return err
	}

	elasticIP, err := elasticIPResp.FindElasticIP(c.ElasticIP)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	var updateReverseDNSElasticIPReq v3.UpdateReverseDNSElasticIPRequest
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.ReverseDNS)) {
		updateReverseDNSElasticIPReq.DomainName = c.ReverseDNS
		updatedRDNS = true
	}

	var updateElasticIPReq v3.UpdateElasticIPRequest

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		updateElasticIPReq.Description = c.Description
		updatedInstance = true
	}

	updatedElasticIPHealthcheck := elasticIP.Healthcheck

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckMode)) {
		updatedElasticIPHealthcheck.Mode = v3.ElasticIPHealthcheckMode(c.HealthcheckMode)
		updatedInstance = true
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
		updatedElasticIPHealthcheck.Interval = c.HealthcheckInterval
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckPort); cmd.Flags().Changed(flag) {
		updatedElasticIPHealthcheck.Port = c.HealthcheckPort
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckStrikesFail); cmd.Flags().Changed(flag) {
		updatedElasticIPHealthcheck.StrikesFail = c.HealthcheckStrikesFail
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckStrikesOK); cmd.Flags().Changed(flag) {
		updatedElasticIPHealthcheck.StrikesOk = c.HealthcheckStrikesOK
		updatedInstance = true
	}

	if updatedElasticIPHealthcheck.Mode == v3.ElasticIPHealthcheckModeHttps {
		if flag := mustCLICommandFlagName(c, &c.HealthcheckTLSSSkipVerify); cmd.Flags().Changed(flag) {
			updatedElasticIPHealthcheck.TlsSkipVerify = &c.HealthcheckTLSSSkipVerify
			updatedInstance = true
		}

		if flag := mustCLICommandFlagName(c, &c.HealthcheckTLSSNI); cmd.Flags().Changed(flag) {
			updatedElasticIPHealthcheck.TlsSNI = c.HealthcheckTLSSNI
			updatedInstance = true
		}
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckTimeout); cmd.Flags().Changed(flag) {
		updatedElasticIPHealthcheck.Timeout = c.HealthcheckTimeout
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckURI); cmd.Flags().Changed(flag) {
		updatedElasticIPHealthcheck.URI = c.HealthcheckURI
		updatedInstance = true
	}

	if updatedInstance || updatedRDNS {
		err = decorateAsyncOperations(fmt.Sprintf("Updating Elastic IP %s...", c.ElasticIP), func() error {
			if updatedInstance {
				updateElasticIPReq.Healthcheck = updatedElasticIPHealthcheck
				op, err := client.UpdateElasticIP(ctx, elasticIP.ID, updateElasticIPReq)
				if err != nil {
					return fmt.Errorf("exoscale: error while updating the Elastic IP: %w", err)
				}

				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
				if err != nil {
					return fmt.Errorf("exoscale: error while waiting for the Elastic IP update: %w", err)
				}
			}

			if updatedRDNS {
				var op *v3.Operation
				var err error

				currentReverseDNS, err := client.GetReverseDNSElasticIP(ctx, elasticIP.ID)
				if err != nil && !(errors.Is(err, v3.ErrNotFound)) {
					return fmt.Errorf("exoscale: error while getting current Elastic IP reverse DNS: %w", err)
				}

				if c.ReverseDNS == "" && currentReverseDNS != nil {
					op, err = client.DeleteReverseDNSElasticIP(ctx, elasticIP.ID)
				} else if c.ReverseDNS != "" {
					op, err = client.UpdateReverseDNSElasticIP(ctx, elasticIP.ID, updateReverseDNSElasticIPReq)
				} else {
					// c.ReverseDNS is "" and the ElasticIP has currently no ReverseDNS
					// The server throws an error, if we try to delete an non-existing reverse dns
					return nil
				}

				if err != nil {
					return fmt.Errorf("exoscale: error while updating the Elastic IP's reverse DNS: %w", err)
				}

				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
				if err != nil {
					return fmt.Errorf("exoscale: error while waiting for Elastic IP's reverse DNS update: %w", err)
				}

			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&elasticIPShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			ElasticIP:          elasticIP.ID.String(),
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
		return nil
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
