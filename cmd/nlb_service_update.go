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

type nlbServiceUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Service             string `cli-arg:"#" cli-usage:"SERVICE-NAME|ID"`

	Description         string      `cli-usage:"service description"`
	HealthcheckInterval int64       `cli-usage:"service health checking interval in seconds"`
	HealthcheckMode     string      `cli-usage:"service health checking mode (tcp|http|https)"`
	HealthcheckPort     int64       `cli-usage:"service health checking port"`
	HealthcheckRetries  int64       `cli-usage:"service health checking retries"`
	HealthcheckTLSSNI   string      `cli-flag:"healthcheck-tls-sni" cli-usage:"service health checking server name to present with SNI in https mode"`
	HealthcheckTimeout  int64       `cli-usage:"service health checking timeout in seconds"`
	HealthcheckURI      string      `cli-usage:"service health checking URI (required in http(s) mode)"`
	Name                string      `cli-usage:"service name"`
	Port                int64       `cli-usage:"service port"`
	Protocol            string      `cli-usage:"service network protocol (tcp|udp)"`
	Strategy            string      `cli-usage:"load balancing strategy (round-robin|source-hash)"`
	TargetPort          int64       `cli-usage:"port to forward traffic to on target instances"`
	Zone                v3.ZoneName `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceUpdateCmd) cmdAliases() []string { return nil }

func (c *nlbServiceUpdateCmd) cmdShort() string { return "Update a Network Load Balancer service" }

func (c *nlbServiceUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Network Load Balancer service.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbServiceShowOutput{}), ", "))
}

func (c *nlbServiceUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	var updated bool

	nlbs, err := client.ListLoadBalancers(ctx)
	if err != nil {
		return err
	}
	nlb, err := nlbs.FindLoadBalancer(c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	var service *v3.LoadBalancerService
	for _, s := range nlb.Services {
		if c.Service == string(s.ID) || c.Service == s.Name {
			service = &s
		}
	}
	if service == nil {
		return errors.New("service not found")
	}

	svc := v3.UpdateLoadBalancerServiceRequest{
		Healthcheck: service.Healthcheck,
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		svc.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckInterval)) {
		svc.Healthcheck.Interval = c.HealthcheckInterval
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckMode)) {
		svc.Healthcheck.Mode = v3.LoadBalancerServiceHealthcheckMode(c.HealthcheckMode)
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckPort)) {
		svc.Healthcheck.Port = c.HealthcheckPort
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckRetries)) {
		svc.Healthcheck.Retries = c.HealthcheckRetries
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckTimeout)) {
		svc.Healthcheck.Timeout = c.HealthcheckTimeout
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckURI)) {
		svc.Healthcheck.URI = c.HealthcheckURI
		updated = true
	}
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.HealthcheckTLSSNI)) {
		svc.Healthcheck.TlsSNI = c.HealthcheckTLSSNI
		updated = true
	}

	// If mode is is tcp, ensure URI and TLSSNI are not set
	if c.HealthcheckMode == "tcp" && c.HealthcheckTLSSNI != "" {
		return fmt.Errorf("cannot setup healthcheck TLSSNI with TCP mode (current value: %q)", c.HealthcheckTLSSNI)
	}
	if c.HealthcheckMode == "tcp" && c.HealthcheckURI != "" {
		return fmt.Errorf("cannot setup healthcheck URI with TCP mode (current value: %q)", c.HealthcheckURI)
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		svc.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Port)) {
		svc.Port = c.Port
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Protocol)) {
		svc.Protocol = v3.UpdateLoadBalancerServiceRequestProtocol(c.Protocol)
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Strategy)) {
		svc.Strategy = v3.UpdateLoadBalancerServiceRequestStrategy(c.Strategy)
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.TargetPort)) {
		svc.TargetPort = c.TargetPort
		updated = true
	}

	if updated {
		op, err := client.UpdateLoadBalancerService(ctx, nlb.ID, service.ID, svc)
		if err != nil {
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Updating service %q...", c.Service), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}

		if !globalstate.Quiet {
			return (&nlbServiceShowCmd{
				cliCommandSettings:  c.cliCommandSettings,
				NetworkLoadBalancer: nlb.ID.String(),
				Service:             service.ID.String(),
				Zone:                c.Zone,
			}).cmdRun(nil, nil)
		}

	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbServiceCmd, &nlbServiceUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
