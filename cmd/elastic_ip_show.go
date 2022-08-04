package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type elasticIPShowOutput struct {
	ID                       string         `json:"id"`
	IPAddress                string         `json:"ip_address"`
	AddressFamily            string         `json:"address_family"`
	CIDR                     string         `json:"cidr"`
	Description              string         `json:"description"`
	Zone                     string         `json:"zone"`
	Type                     string         `json:"type"`
	HealthcheckMode          *string        `json:"healthcheck_mode,omitempty"`
	HealthcheckPort          *uint16        `json:"healthcheck_port,omitempty"`
	HealthcheckURI           *string        `json:"healthcheck_uri,omitempty"`
	HealthcheckInterval      *time.Duration `json:"healthcheck_interval,omitempty"`
	HealthcheckTimeout       *time.Duration `json:"healthcheck_timeout,omitempty"`
	HealthcheckStrikesOK     *int64         `json:"healthcheck_strikes_ok,omitempty"`
	HealthcheckStrikesFail   *int64         `json:"healthcheck_strikes_fail,omitempty"`
	HealthcheckTLSSNI        *string        `json:"healthcheck_tls_sni,omitempty"`
	HealthcheckTLSSkipVerify *bool          `json:"healthcheck_tls_skip_verify,omitempty"`
}

func (o *elasticIPShowOutput) toJSON() { outputJSON(o) }
func (o *elasticIPShowOutput) toText() { outputText(o) }
func (o *elasticIPShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Elastic IP"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"IP Address", o.IPAddress})
	t.Append([]string{"Address Family", o.AddressFamily})
	t.Append([]string{"CIDR", o.CIDR})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"Type", o.Type})

	if o.Type == "managed" {
		t.Append([]string{"Healthcheck Mode", *o.HealthcheckMode})
		t.Append([]string{"Healthcheck Port", fmt.Sprint(*o.HealthcheckPort)})
		if strings.HasPrefix(*o.HealthcheckMode, "http") {
			t.Append([]string{"Healthcheck URI", *o.HealthcheckURI})
		}
		t.Append([]string{"Healthcheck Interval", fmt.Sprint(o.HealthcheckInterval)})
		t.Append([]string{"Healthcheck Timeout", fmt.Sprint(o.HealthcheckTimeout)})
		t.Append([]string{"Healthcheck Strikes OK", fmt.Sprint(*o.HealthcheckStrikesOK)})
		t.Append([]string{"Healthcheck Strikes Fail", fmt.Sprint(*o.HealthcheckStrikesFail)})
		if *o.HealthcheckMode == "https" {
			t.Append([]string{"Healthcheck TLS SNI", defaultString(o.HealthcheckTLSSNI, "")})
			t.Append([]string{"Healthcheck TLS Skip Verification", fmt.Sprint(defaultBool(o.HealthcheckTLSSkipVerify, false))})
		}
	}
}

type elasticIPShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Zone string `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPShowCmd) cmdAliases() []string { return gShowAlias }

func (c *elasticIPShowCmd) cmdShort() string {
	return "Show an Elastic IP details"
}

func (c *elasticIPShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Elastic IP details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	elasticIP, err := cs.FindElasticIP(ctx, c.Zone, c.ElasticIP)
	if err != nil {
		return err
	}

	out := elasticIPShowOutput{
		ID:            *elasticIP.ID,
		IPAddress:     elasticIP.IPAddress.String(),
		AddressFamily: defaultString(elasticIP.AddressFamily, ""),
		CIDR:          defaultString(elasticIP.CIDR, ""),
		Description:   defaultString(elasticIP.Description, ""),
		Zone:          c.Zone,
		Type:          "manual",
	}

	if elasticIP.Healthcheck != nil {
		out.Type = "managed"
		out.HealthcheckMode = elasticIP.Healthcheck.Mode
		out.HealthcheckPort = elasticIP.Healthcheck.Port
		out.HealthcheckURI = elasticIP.Healthcheck.URI
		out.HealthcheckInterval = elasticIP.Healthcheck.Interval
		out.HealthcheckTimeout = elasticIP.Healthcheck.Timeout
		out.HealthcheckStrikesOK = elasticIP.Healthcheck.StrikesOK
		out.HealthcheckStrikesFail = elasticIP.Healthcheck.StrikesFail
		out.HealthcheckTLSSNI = elasticIP.Healthcheck.TLSSNI
		out.HealthcheckTLSSkipVerify = elasticIP.Healthcheck.TLSSkipVerify
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
