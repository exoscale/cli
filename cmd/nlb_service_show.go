package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type nlbServerStatusShowOutput struct {
	InstanceIP string `json:"instance_ip"`
	Status     string `json:"status"`
}

type nlbServiceHealthcheckShowOutput struct {
	Mode     string        `json:"mode"`
	Port     uint16        `json:"port"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int64         `json:"retries"`
	URI      string        `json:"uri"`
	TLSSNI   string        `json:"tls_sni"`
}

type nlbServiceShowOutput struct {
	ID                string                          `json:"id"`
	Name              string                          `json:"name"`
	Description       string                          `json:"description"`
	InstancePoolID    string                          `json:"instance_pool_id"`
	Protocol          string                          `json:"protocol"`
	Port              uint16                          `json:"port"`
	TargetPort        uint16                          `json:"target_port"`
	Strategy          string                          `json:"strategy"`
	Healthcheck       nlbServiceHealthcheckShowOutput `json:"healthcheck"`
	HealthcheckStatus []nlbServerStatusShowOutput     `json:"healthcheck_status"`
	State             string                          `json:"state"`
}

func (o *nlbServiceShowOutput) toJSON() { output.JSON(o) }
func (o *nlbServiceShowOutput) toText() { output.Text(o) }
func (o *nlbServiceShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"NLB Service"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Instance Pool ID", o.InstancePoolID})
	t.Append([]string{"Protocol", o.Protocol})
	t.Append([]string{"Port", fmt.Sprint(o.Port)})
	t.Append([]string{"Target Port", fmt.Sprint(o.TargetPort)})
	t.Append([]string{"Strategy", o.Strategy})
	t.Append([]string{"Healthcheck Mode", o.Healthcheck.Mode})
	t.Append([]string{"Healthcheck Port", fmt.Sprint(o.Healthcheck.Port)})
	if strings.HasPrefix(o.Healthcheck.Mode, "http") {
		t.Append([]string{"Healthcheck URI", o.Healthcheck.URI})
	}
	t.Append([]string{"Healthcheck Interval", fmt.Sprint(o.Healthcheck.Interval)})
	t.Append([]string{"Healthcheck Timeout", fmt.Sprint(o.Healthcheck.Timeout)})
	t.Append([]string{"Healthcheck Retries", fmt.Sprint(o.Healthcheck.Retries)})
	if o.Healthcheck.Mode == "https" {
		t.Append([]string{"Healthcheck TLS SNI", fmt.Sprint(o.Healthcheck.TLSSNI)})
	}
	t.Append([]string{"Healthcheck Status", func() string {
		if len(o.HealthcheckStatus) > 0 {
			return strings.Join(
				func() []string {
					statuses := make([]string, len(o.HealthcheckStatus))
					for i := range o.HealthcheckStatus {
						statuses[i] = fmt.Sprintf("%s | %s",
							o.HealthcheckStatus[i].InstanceIP,
							o.HealthcheckStatus[i].Status)
					}
					return statuses
				}(),
				"\n")
		}
		return "n/a"
	}()})
	t.Append([]string{"State", o.State})
}

type nlbServiceShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Service             string `cli-arg:"#" cli-usage:"SERVICE-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceShowCmd) cmdAliases() []string { return gShowAlias }

func (c *nlbServiceShowCmd) cmdShort() string { return "Show a Network Load Balancer service details" }

func (c *nlbServiceShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Network Load Balancer service details.

Supported output template annotations: %s`,
		strings.Join(output.output.OutputterTemplateAnnotations(&nlbServiceShowOutput{}), ", "))
}

func (c *nlbServiceShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var svc *egoscale.NetworkLoadBalancerService

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	nlb, err := cs.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	for _, s := range nlb.Services {
		if *s.ID == c.Service || *s.Name == c.Service {
			svc = s
			break
		}
	}
	if svc == nil {
		return errors.New("service not found")
	}

	out := nlbServiceShowOutput{
		ID:             *svc.ID,
		Name:           *svc.Name,
		Description:    utils.DefaultString(svc.Description, ""),
		InstancePoolID: *svc.InstancePoolID,
		Protocol:       *svc.Protocol,
		Port:           *svc.Port,
		TargetPort:     *svc.TargetPort,
		Strategy:       *svc.Strategy,
		State:          *svc.State,

		Healthcheck: nlbServiceHealthcheckShowOutput{
			Mode:     *svc.Healthcheck.Mode,
			Port:     *svc.Healthcheck.Port,
			Interval: *svc.Healthcheck.Interval,
			Timeout:  *svc.Healthcheck.Timeout,
			Retries:  *svc.Healthcheck.Retries,
			URI:      utils.DefaultString(svc.Healthcheck.URI, ""),
			TLSSNI:   utils.DefaultString(svc.Healthcheck.TLSSNI, ""),
		},

		HealthcheckStatus: func() []nlbServerStatusShowOutput {
			statuses := make([]nlbServerStatusShowOutput, len(svc.HealthcheckStatus))
			for i, st := range svc.HealthcheckStatus {
				statuses[i] = nlbServerStatusShowOutput{
					InstanceIP: st.InstanceIP.String(),
					Status:     *st.Status,
				}
			}
			return statuses
		}(),
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbServiceCmd, &nlbServiceShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
