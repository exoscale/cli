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

type nlbServiceUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Service             string `cli-arg:"#" cli-usage:"SERVICE-NAME|ID"`

	Description         string `cli-usage:"service description"`
	HealthcheckInterval int64  `cli-usage:"service health checking interval in seconds"`
	HealthcheckMode     string `cli-usage:"service health checking mode (tcp|http|https)"`
	HealthcheckPort     int64  `cli-usage:"service health checking port"`
	HealthcheckRetries  int64  `cli-usage:"service health checking retries"`
	HealthcheckTLSSNI   string `cli-flag:"healthcheck-tls-sni" cli-usage:"service health checking server name to present with SNI in https mode"`
	HealthcheckTimeout  int64  `cli-usage:"service health checking timeout in seconds"`
	HealthcheckURI      string `cli-usage:"service health checking URI (required in http(s) mode)"`
	Name                string `cli-usage:"service name"`
	Port                int64  `cli-usage:"service port"`
	Protocol            string `cli-usage:"service network protocol (tcp|udp)"`
	Strategy            string `cli-usage:"load balancing strategy (round-robin|source-hash)"`
	TargetPort          int64  `cli-usage:"port to forward traffic to on target instances"`
	Zone                string `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceUpdateCmd) CmdAliases() []string { return nil }

func (c *nlbServiceUpdateCmd) CmdShort() string { return "Update a Network Load Balancer service" }

func (c *nlbServiceUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Network Load Balancer service.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbServiceShowOutput{}), ", "))
}

func (c *nlbServiceUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var (
		service *egoscale.NetworkLoadBalancerService
		updated bool
	)

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	nlb, err := globalstate.EgoscaleClient.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	for _, s := range nlb.Services {
		if *s.ID == c.Service || *s.Name == c.Service {
			service = s
			break
		}
	}
	if service == nil {
		return errors.New("service not found")
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		service.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckInterval)) {
		hcInterval := time.Duration(c.HealthcheckInterval) * time.Second
		service.Healthcheck.Interval = &hcInterval
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckMode)) {
		service.Healthcheck.Mode = &c.HealthcheckMode
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckPort)) {
		hcPort := uint16(c.HealthcheckPort)
		service.Healthcheck.Port = &hcPort
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckRetries)) {
		service.Healthcheck.Retries = &c.HealthcheckRetries
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckTLSSNI)) {
		service.Healthcheck.TLSSNI = &c.HealthcheckTLSSNI
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckTimeout)) {
		hcTimeout := time.Duration(c.HealthcheckTimeout) * time.Second
		service.Healthcheck.Timeout = &hcTimeout
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.HealthcheckURI)) {
		service.Healthcheck.URI = &c.HealthcheckURI
		updated = true
	}

	// If mode is is tcp, ensure URI and TLSSNI are not set
	if *service.Healthcheck.Mode == "tcp" && (utils.DefaultString(service.Healthcheck.TLSSNI, "") != "" || utils.DefaultString(service.Healthcheck.URI, "") != "") {
		service.Healthcheck = &egoscale.NetworkLoadBalancerServiceHealthcheck{
			Interval: service.Healthcheck.Interval,
			Mode:     service.Healthcheck.Mode,
			Port:     service.Healthcheck.Port,
			Retries:  service.Healthcheck.Retries,
			Timeout:  service.Healthcheck.Timeout,
		}
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		service.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Port)) {
		port := uint16(c.Port)
		service.Port = &port
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Protocol)) {
		service.Protocol = &c.Protocol
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Strategy)) {
		service.Strategy = &c.Strategy
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TargetPort)) {
		targetPort := uint16(c.TargetPort)
		service.TargetPort = &targetPort
		updated = true
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Updating service %q...", c.Service), func() {
		if updated {
			if err = globalstate.EgoscaleClient.UpdateNetworkLoadBalancerService(ctx, c.Zone, nlb, service); err != nil {
				return
			}
		}
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
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbServiceCmd, &nlbServiceUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
