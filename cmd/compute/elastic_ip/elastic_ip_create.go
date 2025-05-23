package elastic_ip

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

<<<<<<< Updated upstream:cmd/elastic_ip_create.go
=======
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_create.go
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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

func (c *elasticIPCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *elasticIPCreateCmd) CmdShort() string {
	return "Create an Elastic IP"
}

func (c *elasticIPCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Elastic IP.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

<<<<<<< Updated upstream:cmd/elastic_ip_create.go
func (c *elasticIPCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
=======
func (c *elasticIPCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_create.go

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

<<<<<<< Updated upstream:cmd/elastic_ip_create.go
	decorateAsyncOperation("Creating Elastic IP...", func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
=======
	var err error
	utils.DecorateAsyncOperation("Creating Elastic IP...", func() {
		elasticIP, err = globalstate.EgoscaleClient.CreateElasticIP(ctx, c.Zone, elasticIP)
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_create.go
	})
	if err != nil {
		return err
	}

	return (&elasticIPShowCmd{
<<<<<<< Updated upstream:cmd/elastic_ip_create.go
		cliCommandSettings: c.cliCommandSettings,
		ElasticIP:          op.Reference.ID.String(),
=======
		CliCommandSettings: c.CliCommandSettings,
		ElasticIP:          *elasticIP.ID,
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_create.go
		Zone:               c.Zone,
	}).CmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(elasticIPCmd, &elasticIPCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		HealthcheckInterval:    10,
		HealthcheckStrikesFail: 2,
		HealthcheckStrikesOK:   3,
		HealthcheckTimeout:     3,
	}))
}
