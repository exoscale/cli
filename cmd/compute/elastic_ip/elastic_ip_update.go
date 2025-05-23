package elastic_ip

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type elasticIPUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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

func (c *elasticIPUpdateCmd) CmdAliases() []string { return nil }

func (c *elasticIPUpdateCmd) CmdShort() string {
	return "Update an Elastic IP"
}

func (c *elasticIPUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Compute instance Elastic IP.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updatedInstance, updatedRDNS bool

<<<<<<< Updated upstream:cmd/elastic_ip_update.go
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
=======
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	elasticIP, err := globalstate.EgoscaleClient.FindElasticIP(ctx, c.Zone, c.ElasticIP)
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_update.go
	if err != nil {
		return err
	}

	elasticIps, err := client.ListElasticIPS(ctx)
	if err != nil {
		return err
	}

	elasticIP, err := elasticIps.FindElasticIP(c.ElasticIP)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

<<<<<<< Updated upstream:cmd/elastic_ip_update.go
	req := v3.UpdateElasticIPRequest{
		Description: elasticIP.Description,
		Healthcheck: elasticIP.Healthcheck,
		Labels:      elasticIP.Labels,
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		req.Description = c.Description
		updatedInstance = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckMode)) {
		if req.Healthcheck == nil {
			req.Healthcheck = new(v3.ElasticIPHealthcheck)
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		elasticIP.Description = &c.Description
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckMode)) {
		if elasticIP.Healthcheck == nil {
			elasticIP.Healthcheck = new(egoscale.ElasticIPHealthcheck)
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_update.go
		}
		req.Healthcheck.Mode = v3.ElasticIPHealthcheckMode(c.HealthcheckMode)
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ReverseDNS)) {
		updatedRDNS = true
	}

	for _, flag := range []string{
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckInterval),
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckPort),
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckStrikesFail),
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckStrikesOK),
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckTLSSNI),
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckTLSSSkipVerify),
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckTimeout),
		exocmd.MustCLICommandFlagName(c, &c.HealthcheckURI),
	} {
		if cmd.Flags().Changed(flag) && req.Healthcheck == nil {
			return fmt.Errorf("--%s cannot be used on a non-managed Elastic IP", flag)
		}
	}

<<<<<<< Updated upstream:cmd/elastic_ip_update.go
	if flag := mustCLICommandFlagName(c, &c.HealthcheckInterval); cmd.Flags().Changed(flag) {
		req.Healthcheck.Interval = c.HealthcheckInterval
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckPort); cmd.Flags().Changed(flag) {
		req.Healthcheck.Port = c.HealthcheckPort
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckStrikesFail); cmd.Flags().Changed(flag) {
		req.Healthcheck.StrikesFail = c.HealthcheckStrikesFail
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckStrikesOK); cmd.Flags().Changed(flag) {
		req.Healthcheck.StrikesOk = c.HealthcheckStrikesOK
		updatedInstance = true
	}

	if req.Healthcheck != nil && req.Healthcheck.Mode == "https" {
		if flag := mustCLICommandFlagName(c, &c.HealthcheckTLSSSkipVerify); cmd.Flags().Changed(flag) {
			req.Healthcheck.TlsSkipVerify = &c.HealthcheckTLSSSkipVerify
			updatedInstance = true
		}

		if flag := mustCLICommandFlagName(c, &c.HealthcheckTLSSNI); cmd.Flags().Changed(flag) {
			req.Healthcheck.TlsSNI = c.HealthcheckTLSSNI
=======
	if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckInterval); cmd.Flags().Changed(flag) {
		interval := time.Duration(c.HealthcheckInterval) * time.Second
		elasticIP.Healthcheck.Interval = &interval
		updatedInstance = true
	}

	if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckPort); cmd.Flags().Changed(flag) {
		port := uint16(c.HealthcheckPort)
		elasticIP.Healthcheck.Port = &port
		updatedInstance = true
	}

	if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckStrikesFail); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.StrikesFail = &c.HealthcheckStrikesFail
		updatedInstance = true
	}

	if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckStrikesOK); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.StrikesOK = &c.HealthcheckStrikesOK
		updatedInstance = true
	}

	if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckStrikesOK); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.StrikesOK = &c.HealthcheckStrikesOK
		updatedInstance = true
	}

	if elasticIP.Healthcheck != nil && *elasticIP.Healthcheck.Mode == "https" {
		if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckTLSSSkipVerify); cmd.Flags().Changed(flag) {
			elasticIP.Healthcheck.TLSSkipVerify = &c.HealthcheckTLSSSkipVerify
			updatedInstance = true
		}

		if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckTLSSNI); cmd.Flags().Changed(flag) {
			elasticIP.Healthcheck.TLSSNI = &c.HealthcheckTLSSNI
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_update.go
			updatedInstance = true
		}
	}

<<<<<<< Updated upstream:cmd/elastic_ip_update.go
	if flag := mustCLICommandFlagName(c, &c.HealthcheckTimeout); cmd.Flags().Changed(flag) {
		req.Healthcheck.Timeout = c.HealthcheckTimeout
		updatedInstance = true
	}

	if flag := mustCLICommandFlagName(c, &c.HealthcheckURI); cmd.Flags().Changed(flag) {
		req.Healthcheck.URI = c.HealthcheckURI
		updatedInstance = true
	}

	if updatedInstance {
		op, err := client.UpdateElasticIP(ctx, elasticIP.ID, req)
		if err != nil {
			return err
		}
		decorateAsyncOperation(fmt.Sprintf("Updating Elastic IP %s...", c.ElasticIP), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
=======
	if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckTimeout); cmd.Flags().Changed(flag) {
		timeout := time.Duration(c.HealthcheckTimeout) * time.Second
		elasticIP.Healthcheck.Timeout = &timeout
		updatedInstance = true
	}

	if flag := exocmd.MustCLICommandFlagName(c, &c.HealthcheckURI); cmd.Flags().Changed(flag) {
		elasticIP.Healthcheck.URI = &c.HealthcheckURI
		updatedInstance = true
	}

	if updatedInstance || updatedRDNS {
		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Elastic IP %s...", c.ElasticIP), func() {
			if updatedInstance {
				if err = globalstate.EgoscaleClient.UpdateElasticIP(ctx, c.Zone, elasticIP); err != nil {
					return
				}
			}

			if updatedRDNS {
				if c.ReverseDNS == "" {
					err = globalstate.EgoscaleClient.DeleteElasticIPReverseDNS(ctx, c.Zone, *elasticIP.ID)
				} else {
					err = globalstate.EgoscaleClient.UpdateElasticIPReverseDNS(ctx, c.Zone, *elasticIP.ID, c.ReverseDNS)
				}
			}
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_update.go
		})
		if err != nil {
			return err
		}
	}

	if updatedRDNS {
		var op *v3.Operation
		if c.ReverseDNS == "" {
			op, err = client.DeleteReverseDNSElasticIP(ctx, elasticIP.ID)
		} else {
			op, err = client.UpdateReverseDNSElasticIP(ctx, elasticIP.ID, v3.UpdateReverseDNSElasticIPRequest{DomainName: c.ReverseDNS})
		}
		if err != nil {
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Updating Reverse DNS for Elastic IP %s...", c.ElasticIP), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&elasticIPShowCmd{
<<<<<<< Updated upstream:cmd/elastic_ip_update.go
			cliCommandSettings: c.cliCommandSettings,
			ElasticIP:          elasticIP.ID.String(),
=======
			CliCommandSettings: c.CliCommandSettings,
			ElasticIP:          *elasticIP.ID,
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_update.go
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(elasticIPCmd, &elasticIPUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
