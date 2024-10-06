package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPShowOutput struct {
	ID                       v3.UUID                   `json:"id"`
	IPAddress                string                    `json:"ip_address"`
	AddressFamily            v3.ElasticIPAddressfamily `json:"address_family"`
	CIDR                     string                    `json:"cidr"`
	Description              string                    `json:"description"`
	Zone                     v3.ZoneName               `json:"zone"`
	Type                     string                    `json:"type"`
	ReverseDNS               v3.DomainName             `json:"reverse_dns"`
	Instances                []string                  `json:"instances"`
	HealthcheckMode          *string                   `json:"healthcheck_mode,omitempty"`
	HealthcheckPort          *int64                    `json:"healthcheck_port,omitempty"`
	HealthcheckURI           *string                   `json:"healthcheck_uri,omitempty"`
	HealthcheckInterval      *int64                    `json:"healthcheck_interval,omitempty"`
	HealthcheckTimeout       *int64                    `json:"healthcheck_timeout,omitempty"`
	HealthcheckStrikesOK     *int64                    `json:"healthcheck_strikes_ok,omitempty"`
	HealthcheckStrikesFail   *int64                    `json:"healthcheck_strikes_fail,omitempty"`
	HealthcheckTLSSNI        *string                   `json:"healthcheck_tls_sni,omitempty"`
	HealthcheckTLSSkipVerify *bool                     `json:"healthcheck_tls_skip_verify,omitempty"`
}

func (o *elasticIPShowOutput) ToJSON() { output.JSON(o) }
func (o *elasticIPShowOutput) ToText() { output.Text(o) }
func (o *elasticIPShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Elastic IP"})
	defer t.Render()

	t.Append([]string{"ID", o.ID.String()})
	t.Append([]string{"IP Address", o.IPAddress})
	t.Append([]string{"Address Family", string(o.AddressFamily)})
	t.Append([]string{"CIDR", o.CIDR})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", string(o.Zone)})
	t.Append([]string{"Type", o.Type})
	t.Append([]string{"Reverse DNS", string(o.ReverseDNS)})

	instances := ""
	for _, instance := range o.Instances {
		instances += instance + " "
	}
	t.Append([]string{"Instances", instances})

	if o.Type == "managed" {
		t.Append([]string{"Healthcheck Mode", *o.HealthcheckMode})
		t.Append([]string{"Healthcheck Port", fmt.Sprint(*o.HealthcheckPort)})
		if strings.HasPrefix(*o.HealthcheckMode, "http") {
			t.Append([]string{"Healthcheck URI", *o.HealthcheckURI})
		}
		t.Append([]string{"Healthcheck Interval", fmt.Sprint(*o.HealthcheckInterval)})
		t.Append([]string{"Healthcheck Timeout", fmt.Sprint(*o.HealthcheckTimeout)})
		t.Append([]string{"Healthcheck Strikes OK", fmt.Sprint(*o.HealthcheckStrikesOK)})
		t.Append([]string{"Healthcheck Strikes Fail", fmt.Sprint(*o.HealthcheckStrikesFail)})
		if *o.HealthcheckMode == "https" {
			t.Append([]string{"Healthcheck TLS SNI", utils.DefaultString(o.HealthcheckTLSSNI, "")})
			t.Append([]string{"Healthcheck TLS Skip Verification", fmt.Sprint(utils.DefaultBool(o.HealthcheckTLSSkipVerify, false))})
		}
	}
}

type elasticIPShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ElasticIP string `cli-arg:"#" cli-usage:"IP-ADDRESS|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"Elastic IP zone"`
}

func (c *elasticIPShowCmd) cmdAliases() []string { return gShowAlias }

func (c *elasticIPShowCmd) cmdShort() string {
	return "Show an Elastic IP details"
}

func (c *elasticIPShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Elastic IP details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPShowOutput{}), ", "))
}

func (c *elasticIPShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	elasticIPs, err := client.ListElasticIPS(ctx)
	if err != nil {
		return err
	}

	elasticIP, err := elasticIPs.FindElasticIP(c.ElasticIP)
	if err != nil {
		return err
	}

	out := elasticIPShowOutput{
		ID:            elasticIP.ID,
		IPAddress:     elasticIP.IP,
		AddressFamily: elasticIP.Addressfamily,
		CIDR:          elasticIP.Cidr,
		Description:   elasticIP.Description,
		Zone:          c.Zone,
		Type:          "manual",
	}

	rdns, err := client.GetReverseDNSElasticIP(ctx, elasticIP.ID)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			out.ReverseDNS = ""
		} else {
			return err
		}
	} else {
		out.ReverseDNS = rdns.DomainName
	}

	attachedInstances, err := utils.GetInstancesAttachedToEIP(ctx, client, elasticIP.IP)
	if err != nil {
		return err
	}

	for _, instance := range attachedInstances.Instances {
		out.Instances = append(out.Instances, instance.Name)
	}

	if elasticIP.Healthcheck != nil {
		out.Type = "managed"
		// Message for reviewer:
		// I don't really like to use this function for only one this parameter, if you have a better solution, please feel free to suggest it
		// TODO: remove comment before merging
		out.HealthcheckMode = utils.NonEmptyStringPtr(string(elasticIP.Healthcheck.Mode))
		out.HealthcheckPort = &elasticIP.Healthcheck.Port
		out.HealthcheckURI = &elasticIP.Healthcheck.URI
		out.HealthcheckInterval = &elasticIP.Healthcheck.Interval
		out.HealthcheckTimeout = &elasticIP.Healthcheck.Timeout
		out.HealthcheckStrikesOK = &elasticIP.Healthcheck.StrikesOk
		out.HealthcheckStrikesFail = &elasticIP.Healthcheck.StrikesFail
		out.HealthcheckTLSSNI = &elasticIP.Healthcheck.TlsSNI
		out.HealthcheckTLSSkipVerify = elasticIP.Healthcheck.TlsSkipVerify
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
