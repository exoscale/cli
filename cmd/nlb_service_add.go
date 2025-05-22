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

type nlbServiceAddCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Name                string `cli-arg:"#" cli-usage:"SERVICE-NAME"`

	Description         string      `cli-usage:"service description"`
	HealthcheckInterval int64       `cli-usage:"service health checking interval in seconds"`
	HealthcheckMode     string      `cli-usage:"service health checking mode (tcp|http|https)"`
	HealthcheckPort     int64       `cli-usage:"service health checking port (defaults to target port)"`
	HealthcheckRetries  int64       `cli-usage:"service health checking retries"`
	HealthcheckTLSSNI   string      `cli-flag:"healthcheck-tls-sni" cli-usage:"service health checking server name to present with SNI in https mode"`
	HealthcheckTimeout  int64       `cli-usage:"service health checking timeout in seconds"`
	HealthcheckURI      string      `cli-usage:"service health checking URI (required in http(s) mode)"`
	InstancePool        string      `cli-usage:"name or ID of the Instance Pool to forward traffic to"`
	Port                int64       `cli-usage:"service port"`
	Protocol            string      `cli-usage:"service network protocol (tcp|udp)"`
	Strategy            string      `cli-usage:"load balancing strategy (round-robin|source-hash)"`
	TargetPort          int64       `cli-usage:"port to forward traffic to on target instances (defaults to service port)"`
	Zone                v3.ZoneName `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceAddCmd) cmdAliases() []string { return nil }

func (c *nlbServiceAddCmd) cmdShort() string { return "Add a service to a Network Load Balancer" }

func (c *nlbServiceAddCmd) cmdLong() string {
	return fmt.Sprintf(`This command adds a service to a Network Load Balancer.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbServiceShowOutput{}), ", "))
}

func (c *nlbServiceAddCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceAddCmd) cmdRun(_ *cobra.Command, _ []string) error {

	service := v3.AddServiceToLoadBalancerRequest{
		Description: c.Description,
		Healthcheck: &v3.LoadBalancerServiceHealthcheck{
			Interval: c.HealthcheckInterval,
			Mode:     v3.LoadBalancerServiceHealthcheckMode(c.HealthcheckMode),
			Retries:  c.HealthcheckRetries,
			Timeout:  c.HealthcheckTimeout,
		},
		Name:         c.Name,
		Port:         c.Port,
		Protocol:     v3.AddServiceToLoadBalancerRequestProtocol(c.Protocol),
		Strategy:     v3.AddServiceToLoadBalancerRequestStrategy(c.Strategy),
		TargetPort:   c.TargetPort,
		InstancePool: &v3.InstancePool{},
	}

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if strings.HasPrefix(string(service.Healthcheck.Mode), "http") && c.HealthcheckURI == "" {
		return errors.New(`a healthcheck URI is required in "http(s)" mode`)
	}

	if service.TargetPort == 0 {
		service.TargetPort = service.Port
	}
	if service.Healthcheck.Port == 0 {
		service.Healthcheck.Port = service.TargetPort
	}

	if c.HealthcheckURI != "" {
		service.Healthcheck.URI = c.HealthcheckURI
	}
	if c.HealthcheckTLSSNI != "" {
		service.Healthcheck.TlsSNI = c.HealthcheckTLSSNI
	}

	nlbs, err := client.ListLoadBalancers(ctx)
	if err != nil {
		return err
	}
	nlb, err := nlbs.FindLoadBalancer(c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	instancePools, err := client.ListInstancePools(ctx)
	if err != nil {
		return err
	}

	instancePool, err := instancePools.FindInstancePool(c.InstancePool)
	if err != nil {
		return err
	}
	service.InstancePool.ID = instancePool.ID

	op, err := client.AddServiceToLoadBalancer(ctx, nlb.ID, service)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Adding service %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&nlbServiceShowCmd{
			cliCommandSettings:  c.cliCommandSettings,
			NetworkLoadBalancer: nlb.ID.String(),
			Service:             service.Name,
			Zone:                c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbServiceCmd, &nlbServiceAddCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		HealthcheckInterval: 10,
		HealthcheckMode:     "tcp",
		HealthcheckRetries:  1,
		HealthcheckTimeout:  5,
		Protocol:            "tcp",
		Strategy:            "round-robin",
	}))
}
