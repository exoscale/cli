package load_balancer

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
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbServerStatusShowOutput struct {
	InstanceIP string `json:"instance_ip"`
	Status     string `json:"status"`
}

type nlbServiceHealthcheckShowOutput struct {
	Mode     string        `json:"mode"`
	Port     int64         `json:"port"`
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
	Port              int64                           `json:"port"`
	TargetPort        int64                           `json:"target_port"`
	Strategy          string                          `json:"strategy"`
	Healthcheck       nlbServiceHealthcheckShowOutput `json:"healthcheck"`
	HealthcheckStatus []nlbServerStatusShowOutput     `json:"healthcheck_status"`
	State             string                          `json:"state"`
}

func (o *nlbServiceShowOutput) ToJSON() { output.JSON(o) }
func (o *nlbServiceShowOutput) ToText() { output.Text(o) }
func (o *nlbServiceShowOutput) ToTable() {
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
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Service             string `cli-arg:"#" cli-usage:"SERVICE-NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *nlbServiceShowCmd) CmdShort() string { return "Show a Network Load Balancer service details" }

func (c *nlbServiceShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Network Load Balancer service details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbServiceShowOutput{}), ", "))
}

func (c *nlbServiceShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	// var svc *egoscale.NetworkLoadBalancerService

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	nlbs, err := client.ListLoadBalancers(ctx)
	if err != nil {
		return err
	}
	n, err := nlbs.FindLoadBalancer(c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	var svc *v3.LoadBalancerService
	for _, s := range n.Services {
		fmt.Println(c.Service, s.ID.String(), s.Name)
		if c.Service == s.ID.String() || c.Service == s.Name {
			svc = &s
		}
	}
	if svc == nil {
		return errors.New("service not found")
	}

	out := nlbServiceShowOutput{
		ID:             string(svc.ID),
		Name:           svc.Name,
		Description:    svc.Description,
		InstancePoolID: string(svc.InstancePool.ID),
		Protocol:       string(svc.Protocol),
		Port:           svc.Port,
		TargetPort:     svc.TargetPort,
		Strategy:       string(svc.Strategy),
		State:          string(svc.State),

		Healthcheck: nlbServiceHealthcheckShowOutput{
			Mode:     string(svc.Healthcheck.Mode),
			Port:     svc.Healthcheck.Port,
			Interval: time.Duration(svc.Healthcheck.Interval * int64(time.Second)),
			Timeout:  time.Duration(svc.Healthcheck.Timeout * int64(time.Second)),
			Retries:  svc.Healthcheck.Retries,
			URI:      svc.Healthcheck.URI,
			TLSSNI:   svc.Healthcheck.TlsSNI,
		},

		HealthcheckStatus: func() []nlbServerStatusShowOutput {
			statuses := make([]nlbServerStatusShowOutput, len(svc.HealthcheckStatus))
			for i, st := range svc.HealthcheckStatus {
				statuses[i] = nlbServerStatusShowOutput{
					InstanceIP: st.PublicIP.String(),
					Status:     string(st.Status),
				}
			}
			return statuses
		}(),
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbServiceCmd, &nlbServiceShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
