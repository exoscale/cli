package load_balancer

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type nlbServiceAddCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Name                string `cli-arg:"#" cli-usage:"SERVICE-NAME"`

	Description         string `cli-usage:"service description"`
	HealthcheckInterval int64  `cli-usage:"service health checking interval in seconds"`
	HealthcheckMode     string `cli-usage:"service health checking mode (tcp|http|https)"`
	HealthcheckPort     int64  `cli-usage:"service health checking port (defaults to target port)"`
	HealthcheckRetries  int64  `cli-usage:"service health checking retries"`
	HealthcheckTLSSNI   string `cli-flag:"healthcheck-tls-sni" cli-usage:"service health checking server name to present with SNI in https mode"`
	HealthcheckTimeout  int64  `cli-usage:"service health checking timeout in seconds"`
	HealthcheckURI      string `cli-usage:"service health checking URI (required in http(s) mode)"`
	InstancePool        string `cli-usage:"name or ID of the Instance Pool to forward traffic to"`
	Port                int64  `cli-usage:"service port"`
	Protocol            string `cli-usage:"service network protocol (tcp|udp)"`
	Strategy            string `cli-usage:"load balancing strategy (round-robin|source-hash)"`
	TargetPort          int64  `cli-usage:"port to forward traffic to on target instances (defaults to service port)"`
	Zone                string `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceAddCmd) CmdAliases() []string { return nil }

func (c *nlbServiceAddCmd) CmdShort() string { return "Add a service to a Network Load Balancer" }

func (c *nlbServiceAddCmd) CmdLong() string {
	return fmt.Sprintf(`This command adds a service to a Network Load Balancer.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbServiceShowOutput{}), ", "))
}

func (c *nlbServiceAddCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceAddCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var (
		port       = uint16(c.Port)
		targetPort = uint16(c.TargetPort)
		hcPort     = uint16(c.HealthcheckPort)
		hcInterval = time.Duration(c.HealthcheckInterval) * time.Second
		hcTimeout  = time.Duration(c.HealthcheckTimeout) * time.Second
	)

	service := &egoscale.NetworkLoadBalancerService{
		Description: utils.NonEmptyStringPtr(c.Description),
		Healthcheck: &egoscale.NetworkLoadBalancerServiceHealthcheck{
			Interval: &hcInterval,
			Mode:     &c.HealthcheckMode,
			Port:     &hcPort,
			Retries:  &c.HealthcheckRetries,
			TLSSNI:   utils.NonEmptyStringPtr(c.HealthcheckTLSSNI),
			Timeout:  &hcTimeout,
			URI:      utils.NonEmptyStringPtr(c.HealthcheckURI),
		},
		Name:       &c.Name,
		Port:       &port,
		Protocol:   &c.Protocol,
		Strategy:   &c.Strategy,
		TargetPort: &targetPort,
	}

	if strings.HasPrefix(*service.Healthcheck.Mode, "http") && *service.Healthcheck.URI == "" {
		return errors.New(`an healthcheck URI is required in "http(s)" mode`)
	}

	if *service.TargetPort == 0 {
		service.TargetPort = service.Port
	}
	if *service.Healthcheck.Port == 0 {
		service.Healthcheck.Port = service.TargetPort
	}

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	nlb, err := globalstate.EgoscaleClient.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return fmt.Errorf("error retrieving Network Load Balancer: %w", err)
	}

	instancePool, err := globalstate.EgoscaleClient.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		return fmt.Errorf("error retrieving Instance Pool: %w", err)
	}
	service.InstancePoolID = instancePool.ID

	utils.DecorateAsyncOperation(fmt.Sprintf("Adding service %q...", c.Name), func() {
		service, err = globalstate.EgoscaleClient.CreateNetworkLoadBalancerService(ctx, c.Zone, nlb, service)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&nlbServiceShowCmd{
			CliCommandSettings:  c.CliCommandSettings,
			NetworkLoadBalancer: *nlb.ID,
			Service:             *service.ID,
			Zone:                c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbServiceCmd, &nlbServiceAddCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		HealthcheckInterval: 10,
		HealthcheckMode:     "tcp",
		HealthcheckRetries:  1,
		HealthcheckTimeout:  5,
		Protocol:            "tcp",
		Strategy:            "round-robin",
	}))
}
