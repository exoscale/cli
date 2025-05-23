package elastic_ip

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPShowOutput struct {
	ID                       string        `json:"id"`
	IPAddress                string        `json:"ip_address"`
	AddressFamily            string        `json:"address_family"`
	CIDR                     string        `json:"cidr"`
	Description              string        `json:"description"`
	Zone                     string        `json:"zone"`
	Type                     string        `json:"type"`
	ReverseDNS               string        `json:"reverse_dns"`
	Instances                []string      `json:"instances"`
	HealthcheckMode          string        `json:"healthcheck_mode,omitempty"`
	HealthcheckPort          int64         `json:"healthcheck_port,omitempty"`
	HealthcheckURI           string        `json:"healthcheck_uri,omitempty"`
	HealthcheckInterval      time.Duration `json:"healthcheck_interval,omitempty"`
	HealthcheckTimeout       time.Duration `json:"healthcheck_timeout,omitempty"`
	HealthcheckStrikesOK     int64         `json:"healthcheck_strikes_ok,omitempty"`
	HealthcheckStrikesFail   int64         `json:"healthcheck_strikes_fail,omitempty"`
	HealthcheckTLSSNI        string        `json:"healthcheck_tls_sni,omitempty"`
	HealthcheckTLSSkipVerify *bool         `json:"healthcheck_tls_skip_verify,omitempty"`
}

func (o *elasticIPShowOutput) ToJSON() { output.JSON(o) }
func (o *elasticIPShowOutput) ToText() { output.Text(o) }
func (o *elasticIPShowOutput) ToTable() {
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
	t.Append([]string{"Reverse DNS", o.ReverseDNS})

	instances := ""
	for _, instance := range o.Instances {
		instances += instance + " "
	}
	t.Append([]string{"Instances", instances})

	if o.Type == "managed" {
		t.Append([]string{"Healthcheck Mode", o.HealthcheckMode})
		t.Append([]string{"Healthcheck Port", fmt.Sprint(o.HealthcheckPort)})
		if strings.HasPrefix(o.HealthcheckMode, "http") {
			t.Append([]string{"Healthcheck URI", o.HealthcheckURI})
		}
		t.Append([]string{"Healthcheck Interval", fmt.Sprint(o.HealthcheckInterval)})
		t.Append([]string{"Healthcheck Timeout", fmt.Sprint(o.HealthcheckTimeout)})
		t.Append([]string{"Healthcheck Strikes OK", fmt.Sprint(o.HealthcheckStrikesOK)})
		t.Append([]string{"Healthcheck Strikes Fail", fmt.Sprint(o.HealthcheckStrikesFail)})
		if o.HealthcheckMode == "https" {
			t.Append([]string{"Healthcheck TLS SNI", o.HealthcheckTLSSNI})
			t.Append([]string{"Healthcheck TLS Skip Verification", fmt.Sprint(o.HealthcheckTLSSkipVerify)})
		}
	}
}

type elasticIPShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Zone string `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *elasticIPShowCmd) CmdShort() string {
	return "Show an Elastic IP details"
}

func (c *elasticIPShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Elastic IP details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	eips, err := client.ListElasticIPS(ctx)
	if err != nil {
		return err
	}

	elasticIp, err := eips.FindElasticIP(c.ElasticIP)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	out := elasticIPShowOutput{
		ID:            elasticIp.ID.String(),
		IPAddress:     elasticIp.IP,
		AddressFamily: string(elasticIp.Addressfamily),
		CIDR:          elasticIp.Cidr,
		Description:   elasticIp.Description,
		Zone:          c.Zone,
		Type:          "manual",
	}

	rdns, err := client.GetReverseDNSElasticIP(ctx, elasticIp.ID)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			out.ReverseDNS = ""
		} else {
			return err
		}
	}

	if rdns != nil {
		out.ReverseDNS = string(rdns.DomainName)
	}

	attachedInstances, err := utils.GetInstancesAttachedToEIP(ctx, client, elasticIp.IP)
	if err != nil {
		return err
	}

	for _, instance := range attachedInstances {
		out.Instances = append(out.Instances, instance.Name)
	}

	if elasticIp.Healthcheck != nil {
		out.Type = "managed"
		out.HealthcheckMode = string(elasticIp.Healthcheck.Mode)
		out.HealthcheckPort = elasticIp.Healthcheck.Port
		out.HealthcheckURI = elasticIp.Healthcheck.URI
		out.HealthcheckInterval = time.Duration(elasticIp.Healthcheck.Interval)
		out.HealthcheckTimeout = time.Duration(elasticIp.Healthcheck.Timeout)
		out.HealthcheckStrikesOK = elasticIp.Healthcheck.StrikesOk
		out.HealthcheckStrikesFail = elasticIp.Healthcheck.StrikesFail
		out.HealthcheckTLSSNI = elasticIp.Healthcheck.TlsSNI
		out.HealthcheckTLSSkipVerify = elasticIp.Healthcheck.TlsSkipVerify
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(elasticIPCmd, &elasticIPShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
